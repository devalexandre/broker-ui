package models

import "github.com/devalexandre/broker-ui/internal/messaging"

// Server represents a messaging server configuration
type Server struct {
	ID           int
	Name         string
	URL          string
	ProviderType messaging.ProviderType
}

// Topic represents a publisher topic
type Topic struct {
	ID        int
	ServerID  int
	TopicName string
}

// Subscription represents a subscriber configuration
type Subscription struct {
	ID             int
	ServerID       int
	SubName        string
	SubjectPattern string
}

// Message represents a message received from external sources for audit purposes
// Only incoming messages are stored, not messages sent by this Broker UI instance
type Message struct {
	ID           int    `json:"id"`
	ServerID     int    `json:"server_id"`
	Subject      string `json:"subject"`
	Payload      string `json:"payload"`
	ProviderType string `json:"provider_type"`
	Direction    string `json:"direction"` // "received" (sent messages are not stored)
	Timestamp    int64  `json:"timestamp"`
	CreatedAt    string `json:"created_at"`
}
