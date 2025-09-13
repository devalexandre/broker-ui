package providers

import (
	"fmt"
	"strings"

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
	case messaging.ProviderPubSub:
		return NewPubSubProvider(), nil
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
		messaging.ProviderPubSub,
		// messaging.ProviderKafka,    // TODO: Uncomment when implemented
		// messaging.ProviderRedis,    // TODO: Uncomment when implemented
	}
}

// DetectProviderFromURL attempts to detect the provider type from a URL
func (f *Factory) DetectProviderFromURL(url string) messaging.ProviderType {
	url = strings.ToLower(url)

	// Check for specific protocols
	if strings.HasPrefix(url, "nats://") || strings.Contains(url, ":4222") {
		return messaging.ProviderNATS
	}

	if strings.HasPrefix(url, "amqp://") || strings.HasPrefix(url, "amqps://") || strings.Contains(url, ":5672") {
		return messaging.ProviderRabbitMQ
	}

	// Check for Pub/Sub patterns
	if strings.Contains(url, ":8085") || strings.Contains(url, "pubsub") || strings.HasPrefix(url, "gcp://") {
		return messaging.ProviderPubSub
	}

	// Check if it looks like a GCP project ID (no special chars, reasonable length)
	if !strings.Contains(url, ":") && !strings.Contains(url, "/") && len(url) > 3 && len(url) < 64 {
		return messaging.ProviderPubSub
	}

	// Default to NATS if no specific pattern is detected
	return messaging.ProviderNATS
}
