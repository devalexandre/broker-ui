package providers

import (
	"fmt"

	"github.com/devalexandre/broker-ui/internal/messaging"
)

// Factory implements messaging.ProviderFactory
type Factory struct{}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProvider creates a messaging provider based on the provider type
func (f *Factory) CreateProvider(providerType messaging.ProviderType) (messaging.MessagingProvider, error) {
	switch providerType {
	case messaging.ProviderNATS:
		return NewNATSProvider(), nil
	case messaging.ProviderRabbitMQ:
		return NewRabbitMQProvider(), nil
	case messaging.ProviderKafka:
		// TODO: Implement Kafka provider
		return nil, fmt.Errorf("Kafka provider not implemented yet")
	case messaging.ProviderRedis:
		// TODO: Implement Redis provider
		return nil, fmt.Errorf("Redis provider not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

// GetSupportedProviders returns a list of supported provider types
func (f *Factory) GetSupportedProviders() []messaging.ProviderType {
	return []messaging.ProviderType{
		messaging.ProviderNATS,
		messaging.ProviderRabbitMQ,
		// messaging.ProviderKafka,    // TODO: Uncomment when implemented
		// messaging.ProviderRedis,    // TODO: Uncomment when implemented
	}
}
