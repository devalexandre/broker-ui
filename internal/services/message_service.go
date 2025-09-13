package services

import (
	"fmt"
	"log"
	"sync"

	"github.com/devalexandre/broker-ui/internal/database"
	"github.com/devalexandre/broker-ui/internal/messaging"
)

type MessageService struct {
	topicRepo        *database.TopicRepository
	subscriptionRepo *database.SubscriptionRepository
	sentMessages     map[string][]string
	receivedMessages map[string][]string
	dashboardCounts  map[string]int
	mutex            sync.RWMutex
}

// NewMessageService creates a new message service
func NewMessageService(topicRepo *database.TopicRepository, subscriptionRepo *database.SubscriptionRepository) *MessageService {
	return &MessageService{
		topicRepo:        topicRepo,
		subscriptionRepo: subscriptionRepo,
		sentMessages:     make(map[string][]string),
		receivedMessages: make(map[string][]string),
		dashboardCounts:  make(map[string]int),
	}
}

// SaveTopic saves a new topic
func (s *MessageService) SaveTopic(serverID int, topicName string) error {
	return s.topicRepo.Save(serverID, topicName)
}

// DeleteTopic deletes a topic
func (s *MessageService) DeleteTopic(topicName string, serverID int) error {
	return s.topicRepo.Delete(topicName, serverID)
}

// SaveSubscription saves a new subscription
func (s *MessageService) SaveSubscription(serverID int, subName, subjectPattern string) error {
	return s.subscriptionRepo.Save(serverID, subName, subjectPattern)
}

// DeleteSubscription deletes a subscription
func (s *MessageService) DeleteSubscription(subName string, serverID int) error {
	return s.subscriptionRepo.Delete(subName, serverID)
}

// PublishMessage publishes a message to a topic
func (s *MessageService) PublishMessage(provider messaging.MessagingProvider, subject, payload string) error {
	if payload == "" {
		return nil
	}

	err := provider.Publish(subject, []byte(payload))
	if err != nil {
		return fmt.Errorf("error publishing message: %w", err)
	}

	log.Printf("Sending message to topic %s: %s", subject, payload)

	// Store sent message
	s.mutex.Lock()
	s.sentMessages[subject] = append(s.sentMessages[subject], payload)
	s.mutex.Unlock()

	return nil
}

// Subscribe subscribes to a subject pattern
func (s *MessageService) Subscribe(provider messaging.MessagingProvider, subName, subjectPattern string, messageChan chan<- string) error {
	s.mutex.Lock()
	s.receivedMessages[subName] = []string{}
	s.mutex.Unlock()

	return provider.Subscribe(subjectPattern, func(subject string, data []byte) {
		payload := string(data)
		log.Printf("Received message from sub %s (subject: %s): %s", subName, subject, payload)

		message := fmt.Sprintf("[%s] %s", subject, payload)

		s.mutex.Lock()
		s.receivedMessages[subName] = append(s.receivedMessages[subName], message)
		s.dashboardCounts[subName]++
		s.mutex.Unlock()

		// Send to channel if provided
		if messageChan != nil {
			select {
			case messageChan <- message:
			default:
				// Channel is full, skip this message
			}
		}
	})
}

// GetSentMessages returns sent messages for a topic
func (s *MessageService) GetSentMessages(topicName string) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.sentMessages[topicName]
}

// GetReceivedMessages returns received messages for a subscription
func (s *MessageService) GetReceivedMessages(subName string) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.receivedMessages[subName]
}

// GetDashboardCounts returns message counts for dashboard
func (s *MessageService) GetDashboardCounts() map[string]int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create a copy to avoid data races
	result := make(map[string]int)
	for k, v := range s.dashboardCounts {
		result[k] = v
	}
	return result
}
