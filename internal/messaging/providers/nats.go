package providers

import (
	"fmt"
	"sync"

	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/nats-io/nats.go"
)

// NATSProvider implements MessagingProvider for NATS
type NATSProvider struct {
	conn          *nats.Conn
	url           string
	subscriptions map[string]*nats.Subscription
	handlers      map[string]messaging.MessageHandler
	mutex         sync.RWMutex
	connected     bool
}

// NewNATSProvider creates a new NATS provider
func NewNATSProvider() *NATSProvider {
	return &NATSProvider{
		subscriptions: make(map[string]*nats.Subscription),
		handlers:      make(map[string]messaging.MessageHandler),
	}
}

// Connect establishes a connection to the NATS server
func (n *NATSProvider) Connect(url string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.connected {
		return nil
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS server at %s: %w", url, err)
	}

	n.conn = conn
	n.url = url
	n.connected = true

	fmt.Printf("Connected to NATS server at %s\n", url)
	return nil
}

// Publish sends a message to the specified subject
func (n *NATSProvider) Publish(subject string, data []byte) error {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	if !n.connected || n.conn == nil {
		return fmt.Errorf("not connected to NATS server")
	}

	err := n.conn.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish message to subject %s: %w", subject, err)
	}

	fmt.Printf("Published message to subject: %s, data: %s\n", subject, string(data))
	return nil
}

// Subscribe subscribes to a subject pattern with a message handler
func (n *NATSProvider) Subscribe(subjectPattern string, handler messaging.MessageHandler) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if !n.connected || n.conn == nil {
		return fmt.Errorf("not connected to NATS server")
	}

	// Check if already subscribed
	if _, exists := n.subscriptions[subjectPattern]; exists {
		return fmt.Errorf("already subscribed to subject pattern: %s", subjectPattern)
	}

	// Create NATS message handler wrapper
	natsHandler := func(msg *nats.Msg) {
		handler(msg.Subject, msg.Data)
	}

	sub, err := n.conn.Subscribe(subjectPattern, natsHandler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to subject pattern %s: %w", subjectPattern, err)
	}

	n.subscriptions[subjectPattern] = sub
	n.handlers[subjectPattern] = handler

	fmt.Printf("Subscribed to subject pattern: %s\n", subjectPattern)
	return nil
}

// Unsubscribe removes a subscription
func (n *NATSProvider) Unsubscribe(subjectPattern string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	sub, exists := n.subscriptions[subjectPattern]
	if !exists {
		return fmt.Errorf("no subscription found for subject pattern: %s", subjectPattern)
	}

	err := sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from subject pattern %s: %w", subjectPattern, err)
	}

	delete(n.subscriptions, subjectPattern)
	delete(n.handlers, subjectPattern)

	fmt.Printf("Unsubscribed from subject pattern: %s\n", subjectPattern)
	return nil
}

// Close closes the connection to the NATS server
func (n *NATSProvider) Close() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if !n.connected || n.conn == nil {
		return nil
	}

	// Unsubscribe from all subscriptions
	for subjectPattern := range n.subscriptions {
		if sub := n.subscriptions[subjectPattern]; sub != nil {
			sub.Unsubscribe()
		}
	}

	n.conn.Close()
	n.connected = false
	n.subscriptions = make(map[string]*nats.Subscription)
	n.handlers = make(map[string]messaging.MessageHandler)

	fmt.Println("Disconnected from NATS server")
	return nil
}

// IsConnected returns true if connected to the NATS server
func (n *NATSProvider) IsConnected() bool {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.connected && n.conn != nil && !n.conn.IsClosed()
}

// GetProviderType returns the provider type
func (n *NATSProvider) GetProviderType() messaging.ProviderType {
	return messaging.ProviderNATS
}
