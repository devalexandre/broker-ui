package messaging

// MessageHandler represents a function that handles incoming messages
type MessageHandler func(subject string, data []byte)

// MessagingProvider defines the interface for messaging systems
type MessagingProvider interface {
	// Connect establishes a connection to the messaging system
	Connect(url string) error

	// Publish sends a message to the specified subject/topic
	Publish(subject string, data []byte) error

	// Subscribe subscribes to a subject/topic pattern with a message handler
	Subscribe(subjectPattern string, handler MessageHandler) error

	// Unsubscribe removes a subscription
	Unsubscribe(subjectPattern string) error

	// Close closes the connection to the messaging system
	Close() error

	// IsConnected returns true if connected to the messaging system
	IsConnected() bool

	// GetProviderType returns the type of messaging provider
	GetProviderType() ProviderType
}

// ProviderType represents different messaging provider types
type ProviderType string

const (
	ProviderNATS     ProviderType = "NATS"
	ProviderRabbitMQ ProviderType = "RabbitMQ"
	ProviderKafka    ProviderType = "Kafka"
	ProviderRedis    ProviderType = "Redis"
)

// ProviderFactory creates messaging providers
type ProviderFactory interface {
	CreateProvider(providerType ProviderType) (MessagingProvider, error)
}

// Message represents a message with metadata
type Message struct {
	Subject   string
	Data      []byte
	Provider  ProviderType
	Timestamp int64
}
