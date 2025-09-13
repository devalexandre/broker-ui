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

// Message represents a message sent or received
type Message struct {
	Subject   string
	Payload   string
	Timestamp string
}
