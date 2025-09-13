package services

import (
	"fmt"

	"github.com/devalexandre/broker-ui/internal/database"
	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/devalexandre/broker-ui/internal/messaging/providers"
	"github.com/devalexandre/broker-ui/internal/models"
)

type ServerService struct {
	serverRepo         *database.ServerRepository
	topicRepo          *database.TopicRepository
	subscriptionRepo   *database.SubscriptionRepository
	messagingProviders map[int]messaging.MessagingProvider
	providerFactory    messaging.ProviderFactory
}

// NewServerService creates a new server service
func NewServerService(serverRepo *database.ServerRepository, topicRepo *database.TopicRepository, subscriptionRepo *database.SubscriptionRepository) *ServerService {
	return &ServerService{
		serverRepo:         serverRepo,
		topicRepo:          topicRepo,
		subscriptionRepo:   subscriptionRepo,
		messagingProviders: make(map[int]messaging.MessagingProvider),
		providerFactory:    providers.NewFactory(),
	}
}

// GetAllServers returns all servers
func (s *ServerService) GetAllServers() ([]models.Server, error) {
	return s.serverRepo.GetAll()
}

// SaveServer saves a new server
func (s *ServerService) SaveServer(name, url string, providerType messaging.ProviderType) error {
	return s.serverRepo.Save(name, url, providerType)
}

// UpdateServer updates an existing server
func (s *ServerService) UpdateServer(serverID int, name, url string, providerType messaging.ProviderType) error {
	return s.serverRepo.Update(serverID, name, url, providerType)
}

// ConnectToServer establishes a connection to a messaging server
func (s *ServerService) ConnectToServer(serverID int, url string, providerType messaging.ProviderType) error {
	provider, err := s.providerFactory.CreateProvider(providerType)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	err = provider.Connect(url)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	s.messagingProviders[serverID] = provider
	return nil
}

// DisconnectFromServer closes the connection to a messaging server
func (s *ServerService) DisconnectFromServer(serverID int) {
	if provider, ok := s.messagingProviders[serverID]; ok {
		provider.Close()
		delete(s.messagingProviders, serverID)
	}
}

// GetMessagingProvider returns the messaging provider for a server
func (s *ServerService) GetMessagingProvider(serverID int) (messaging.MessagingProvider, bool) {
	provider, ok := s.messagingProviders[serverID]
	return provider, ok
}

// GetTopicsForServer returns all topics for a server
func (s *ServerService) GetTopicsForServer(serverID int) ([]models.Topic, error) {
	return s.topicRepo.GetByServerID(serverID)
}

// GetSubscriptionsForServer returns all subscriptions for a server
func (s *ServerService) GetSubscriptionsForServer(serverID int) ([]models.Subscription, error) {
	return s.subscriptionRepo.GetByServerID(serverID)
}
