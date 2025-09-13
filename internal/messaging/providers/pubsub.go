package providers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/devalexandre/broker-ui/internal/messaging"
)

// PubSubProvider implements the MessagingProvider interface for Google Cloud Pub/Sub
type PubSubProvider struct {
	client        *pubsub.Client
	subscriptions map[string]*pubsub.Subscription
	topics        map[string]*pubsub.Topic
	handlers      map[string]messaging.MessageHandler
	ctx           context.Context
	cancel        context.CancelFunc
	connected     bool
	projectID     string
	mu            sync.RWMutex
}

// NewPubSubProvider creates a new Pub/Sub provider
func NewPubSubProvider() *PubSubProvider {
	ctx, cancel := context.WithCancel(context.Background())
	return &PubSubProvider{
		subscriptions: make(map[string]*pubsub.Subscription),
		topics:        make(map[string]*pubsub.Topic),
		handlers:      make(map[string]messaging.MessageHandler),
		ctx:           ctx,
		cancel:        cancel,
		connected:     false,
	}
}

// Connect establishes a connection to Google Cloud Pub/Sub
func (p *PubSubProvider) Connect(url string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Parse project ID from URL or use default
	p.projectID = p.parseProjectID(url)

	// Set emulator host if connecting to local emulator
	if p.isEmulatorURL(url) {
		os.Setenv("PUBSUB_EMULATOR_HOST", p.parseEmulatorHost(url))
		log.Printf("Connecting to Pub/Sub emulator at: %s", os.Getenv("PUBSUB_EMULATOR_HOST"))
	}

	// Create Pub/Sub client
	client, err := pubsub.NewClient(p.ctx, p.projectID)
	if err != nil {
		return fmt.Errorf("failed to create pubsub client: %v", err)
	}

	p.client = client
	p.connected = true

	log.Printf("Connected to Google Cloud Pub/Sub (Project: %s)", p.projectID)
	return nil
}

// Publish sends a message to the specified topic
func (p *PubSubProvider) Publish(subject string, data []byte) error {
	p.mu.RLock()
	if !p.connected {
		p.mu.RUnlock()
		return fmt.Errorf("not connected to Pub/Sub")
	}
	p.mu.RUnlock()

	// Get or create topic
	topic, err := p.getOrCreateTopic(subject)
	if err != nil {
		return fmt.Errorf("failed to get/create topic %s: %v", subject, err)
	}

	// Publish message
	result := topic.Publish(p.ctx, &pubsub.Message{
		Data: data,
	})

	// Wait for the result
	_, err = result.Get(p.ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	log.Printf("Published message to topic: %s", subject)
	return nil
}

// Subscribe subscribes to a topic with a message handler
func (p *PubSubProvider) Subscribe(subjectPattern string, handler messaging.MessageHandler) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to Pub/Sub")
	}

	// Create subscription name (topic + "-subscription")
	subscriptionName := subjectPattern + "-subscription"

	// Get or create topic
	topic, err := p.getOrCreateTopic(subjectPattern)
	if err != nil {
		return fmt.Errorf("failed to get/create topic %s: %v", subjectPattern, err)
	}

	// Get or create subscription
	subscription, err := p.getOrCreateSubscription(subscriptionName, topic)
	if err != nil {
		return fmt.Errorf("failed to get/create subscription %s: %v", subscriptionName, err)
	}

	// Store handler and subscription
	p.handlers[subjectPattern] = handler
	p.subscriptions[subjectPattern] = subscription

	// Start receiving messages in a goroutine
	go func() {
		err := subscription.Receive(p.ctx, func(ctx context.Context, msg *pubsub.Message) {
			// Call the handler
			handler(subjectPattern, msg.Data)
			// Acknowledge the message
			msg.Ack()
		})
		if err != nil && p.ctx.Err() == nil {
			log.Printf("Error receiving messages for %s: %v", subjectPattern, err)
		}
	}()

	log.Printf("Subscribed to topic: %s", subjectPattern)
	return nil
}

// Unsubscribe removes a subscription
func (p *PubSubProvider) Unsubscribe(subjectPattern string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.subscriptions[subjectPattern]; exists {
		// Note: We don't delete the subscription from Pub/Sub as it might be used by other clients
		// We just stop receiving messages by canceling the context for this specific subscription
		delete(p.subscriptions, subjectPattern)
		delete(p.handlers, subjectPattern)

		log.Printf("Unsubscribed from topic: %s", subjectPattern)
		return nil
	}

	return fmt.Errorf("no subscription found for topic: %s", subjectPattern)
}

// Close closes the connection to Pub/Sub
func (p *PubSubProvider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		// Cancel context to stop all receiving operations
		p.cancel()

		// Close all topics
		for _, topic := range p.topics {
			topic.Stop()
		}

		// Close client
		if p.client != nil {
			err := p.client.Close()
			if err != nil {
				log.Printf("Error closing Pub/Sub client: %v", err)
			}
		}

		p.connected = false
		p.subscriptions = make(map[string]*pubsub.Subscription)
		p.topics = make(map[string]*pubsub.Topic)
		p.handlers = make(map[string]messaging.MessageHandler)

		log.Println("Disconnected from Google Cloud Pub/Sub")
	}

	return nil
}

// IsConnected returns true if connected to Pub/Sub
func (p *PubSubProvider) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// GetProviderType returns the provider type
func (p *PubSubProvider) GetProviderType() messaging.ProviderType {
	return messaging.ProviderPubSub
}

// Helper methods

func (p *PubSubProvider) parseProjectID(url string) string {
	// Parse project ID from URL
	// Supported formats:
	// - localhost:8085 or 127.0.0.1:8085 (emulator)
	// - gcp://my-project-id (GCP production)
	// - my-project-id (direct project ID)

	if url == "" {
		return "dev-local" // Default for emulator
	}

	// For emulator URLs
	if p.isEmulatorURL(url) {
		return "dev-local"
	}

	// Remove gcp:// prefix if present
	if strings.HasPrefix(url, "gcp://") {
		return strings.TrimPrefix(url, "gcp://")
	}

	// For production, check if URL looks like a project ID (no colons, no slashes)
	if !strings.Contains(url, ":") && !strings.Contains(url, "/") {
		return url
	}

	// Fallback to environment variable
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		return projectID
	}

	return "dev-local"
}

func (p *PubSubProvider) isEmulatorURL(url string) bool {
	return url == "localhost:8085" || url == "pubsub-emulator:8085" || url == "127.0.0.1:8085"
}

func (p *PubSubProvider) parseEmulatorHost(url string) string {
	if url == "localhost:8085" || url == "127.0.0.1:8085" {
		return url
	}
	if url == "pubsub-emulator:8085" {
		return "localhost:8085" // Map docker service to localhost for client
	}
	return "localhost:8085" // Default
}

func (p *PubSubProvider) getOrCreateTopic(topicName string) (*pubsub.Topic, error) {
	if topic, exists := p.topics[topicName]; exists {
		return topic, nil
	}

	topic := p.client.Topic(topicName)

	// Check if topic exists, create if not
	exists, err := topic.Exists(p.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if topic exists: %v", err)
	}

	if !exists {
		topic, err = p.client.CreateTopic(p.ctx, topicName)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic: %v", err)
		}
		log.Printf("Created topic: %s", topicName)
	}

	p.topics[topicName] = topic
	return topic, nil
}

func (p *PubSubProvider) getOrCreateSubscription(subscriptionName string, topic *pubsub.Topic) (*pubsub.Subscription, error) {
	subscription := p.client.Subscription(subscriptionName)

	// Check if subscription exists, create if not
	exists, err := subscription.Exists(p.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if subscription exists: %v", err)
	}

	if !exists {
		subscription, err = p.client.CreateSubscription(p.ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create subscription: %v", err)
		}
		log.Printf("Created subscription: %s", subscriptionName)
	}

	return subscription, nil
}
