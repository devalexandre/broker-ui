package providers

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/devalexandre/broker-ui/internal/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQProvider implements MessagingProvider for RabbitMQ
type RabbitMQProvider struct {
	url           string
	conn          *amqp.Connection
	channel       *amqp.Channel
	connected     bool
	subscriptions map[string]*amqpSubscription
	mutex         sync.RWMutex
}

type amqpSubscription struct {
	queueName    string
	consumerTag  string
	handler      messaging.MessageHandler
	deliveryChan <-chan amqp.Delivery
	done         chan bool
}

// NewRabbitMQProvider creates a new RabbitMQ provider
func NewRabbitMQProvider() *RabbitMQProvider {
	return &RabbitMQProvider{
		subscriptions: make(map[string]*amqpSubscription),
	}
}

// Connect establishes a connection to the RabbitMQ server
func (r *RabbitMQProvider) Connect(url string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connected {
		return nil
	}

	// Add amqp:// protocol if not present
	connectionURL := url
	if !strings.HasPrefix(url, "amqp://") && !strings.HasPrefix(url, "amqps://") {
		connectionURL = "amqp://" + url
	}

	// Establish connection to RabbitMQ
	conn, err := amqp.Dial(connectionURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.url = connectionURL
	r.conn = conn
	r.channel = ch
	r.connected = true

	log.Printf("Connected to RabbitMQ server at %s", url)

	// Handle connection errors
	go r.handleConnectionErrors()

	return nil
}

// handleConnectionErrors monitors connection and handles reconnection
func (r *RabbitMQProvider) handleConnectionErrors() {
	connErrors := r.conn.NotifyClose(make(chan *amqp.Error))
	channelErrors := r.channel.NotifyClose(make(chan *amqp.Error))

	select {
	case err := <-connErrors:
		if err != nil {
			log.Printf("RabbitMQ connection error: %s", err)
			r.mutex.Lock()
			r.connected = false
			r.mutex.Unlock()
		}
	case err := <-channelErrors:
		if err != nil {
			log.Printf("RabbitMQ channel error: %s", err)
			r.mutex.Lock()
			r.connected = false
			r.mutex.Unlock()
		}
	}
}

// Publish sends a message to the specified exchange/routing key
func (r *RabbitMQProvider) Publish(subject string, data []byte) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.connected || r.channel == nil {
		return fmt.Errorf("not connected to RabbitMQ server")
	}

	// Parse subject as exchange.routingkey or use default exchange
	exchange := ""
	routingKey := subject

	// If subject contains a dot, treat first part as exchange
	if len(subject) > 0 && subject[0] != '.' {
		// For now, use default exchange (empty string) and subject as routing key
		// Later can be enhanced to support exchange.routingkey format
	}

	// Declare queue to ensure it exists (for direct routing)
	_, err := r.channel.QueueDeclare(
		routingKey, // queue name same as routing key for simplicity
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Printf("Warning: failed to declare queue %s: %s", routingKey, err)
	}

	// Publish the message
	err = r.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         data,
			Timestamp:    time.Now(),
			DeliveryMode: amqp.Persistent, // make message persistent
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published message to RabbitMQ queue: %s", routingKey)
	return nil
}

// Subscribe subscribes to a queue/exchange pattern
func (r *RabbitMQProvider) Subscribe(subjectPattern string, handler messaging.MessageHandler) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected || r.channel == nil {
		return fmt.Errorf("not connected to RabbitMQ server")
	}

	// Check if already subscribed
	if _, exists := r.subscriptions[subjectPattern]; exists {
		return fmt.Errorf("already subscribed to queue: %s", subjectPattern)
	}

	// Declare queue
	queue, err := r.channel.QueueDeclare(
		subjectPattern, // queue name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Start consuming
	msgs, err := r.channel.Consume(
		queue.Name, // queue
		"",         // consumer tag (auto-generated)
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Create subscription
	sub := &amqpSubscription{
		queueName:    queue.Name,
		consumerTag:  "", // auto-generated by RabbitMQ
		handler:      handler,
		deliveryChan: msgs,
		done:         make(chan bool),
	}

	r.subscriptions[subjectPattern] = sub

	// Start message processing goroutine
	go r.processMessages(sub, subjectPattern)

	log.Printf("Subscribed to RabbitMQ queue: %s", subjectPattern)
	return nil
}

// processMessages handles incoming messages for a subscription
func (r *RabbitMQProvider) processMessages(sub *amqpSubscription, subjectPattern string) {
	for {
		select {
		case msg, ok := <-sub.deliveryChan:
			if !ok {
				log.Printf("Delivery channel closed for queue: %s", subjectPattern)
				return
			}

			// Call the handler with subject and data
			sub.handler(msg.RoutingKey, msg.Body)

		case <-sub.done:
			log.Printf("Stopping message processing for queue: %s", subjectPattern)
			return
		}
	}
}

// Unsubscribe removes a subscription
func (r *RabbitMQProvider) Unsubscribe(subjectPattern string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	sub, exists := r.subscriptions[subjectPattern]
	if !exists {
		return fmt.Errorf("no subscription found for queue: %s", subjectPattern)
	}

	// Cancel the consumer
	if r.channel != nil && sub.consumerTag != "" {
		err := r.channel.Cancel(sub.consumerTag, false)
		if err != nil {
			log.Printf("Warning: failed to cancel consumer for queue %s: %s", subjectPattern, err)
		}
	}

	// Signal the message processing goroutine to stop
	close(sub.done)

	// Remove from subscriptions
	delete(r.subscriptions, subjectPattern)

	log.Printf("Unsubscribed from RabbitMQ queue: %s", subjectPattern)
	return nil
}

// Close closes the connection to the RabbitMQ server
func (r *RabbitMQProvider) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected {
		return nil
	}

	// Stop all subscriptions
	for pattern, sub := range r.subscriptions {
		if sub.consumerTag != "" && r.channel != nil {
			r.channel.Cancel(sub.consumerTag, false)
		}
		close(sub.done)
		log.Printf("Stopped subscription for queue: %s", pattern)
	}

	// Close channel and connection
	if r.channel != nil {
		r.channel.Close()
		r.channel = nil
	}

	if r.conn != nil {
		r.conn.Close()
		r.conn = nil
	}

	r.connected = false
	r.subscriptions = make(map[string]*amqpSubscription)

	log.Println("Disconnected from RabbitMQ server")
	return nil
}

// IsConnected returns true if connected to the RabbitMQ server
func (r *RabbitMQProvider) IsConnected() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.connected
}

// GetProviderType returns the provider type
func (r *RabbitMQProvider) GetProviderType() messaging.ProviderType {
	return messaging.ProviderRabbitMQ
}
