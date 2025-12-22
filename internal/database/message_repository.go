package database

import (
	"database/sql"
	"time"

	"github.com/devalexandre/broker-ui/internal/models"
)

// MessageRepository handles message audit operations for RECEIVED messages only
//
// This repository is designed to audit messages received from external sources,
// not messages sent by this Broker UI instance. This prevents duplication and
// focuses on tracking external message traffic.
//
// Example usage:
//
//	repo := NewMessageRepository(db)
//
//	// Save a received message (from external source)
//	message := &models.Message{
//	    ServerID:     1,
//	    Subject:      "orders.created",
//	    Payload:      `{"orderId": 123, "amount": 99.99}`,
//	    ProviderType: "NATS",
//	    Direction:    "received",
//	    Timestamp:    time.Now().Unix(),
//	}
//	err := repo.SaveMessage(message)
//
//	// Query recent received messages
//	messages, err := repo.GetRecentMessages(50)
//
//	// Query received messages by server
//	serverMessages, err := repo.GetMessagesByServer(1, 100, 0)
//
//	// Cleanup old messages (older than 30 days)
//	err := repo.DeleteOldMessages(30)
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// SaveMessage saves a message to the database for audit purposes
func (r *MessageRepository) SaveMessage(message *models.Message) error {
	query := `
		INSERT INTO messages (server_id, subject, payload, provider_type, direction, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		message.ServerID,
		message.Subject,
		message.Payload,
		message.ProviderType,
		message.Direction,
		message.Timestamp,
		time.Now().Format(time.RFC3339),
	)

	return err
}

// GetMessagesByServer retrieves all messages for a specific server
func (r *MessageRepository) GetMessagesByServer(serverID int, limit int, offset int) ([]models.Message, error) {
	query := `
		SELECT id, server_id, subject, payload, provider_type, direction, timestamp, created_at
		FROM messages 
		WHERE server_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, serverID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ServerID,
			&msg.Subject,
			&msg.Payload,
			&msg.ProviderType,
			&msg.Direction,
			&msg.Timestamp,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetMessagesBySubject retrieves messages filtered by subject pattern
func (r *MessageRepository) GetMessagesBySubject(serverID int, subjectPattern string, limit int, offset int) ([]models.Message, error) {
	query := `
		SELECT id, server_id, subject, payload, provider_type, direction, timestamp, created_at
		FROM messages 
		WHERE server_id = ? AND subject LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, serverID, "%"+subjectPattern+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ServerID,
			&msg.Subject,
			&msg.Payload,
			&msg.ProviderType,
			&msg.Direction,
			&msg.Timestamp,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetRecentMessages retrieves the most recent messages across all servers
func (r *MessageRepository) GetRecentMessages(limit int) ([]models.Message, error) {
	query := `
		SELECT id, server_id, subject, payload, provider_type, direction, timestamp, created_at
		FROM messages 
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ServerID,
			&msg.Subject,
			&msg.Payload,
			&msg.ProviderType,
			&msg.Direction,
			&msg.Timestamp,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteOldMessages removes messages older than specified days
func (r *MessageRepository) DeleteOldMessages(daysOld int) error {
	query := `
		DELETE FROM messages 
		WHERE created_at < datetime('now', '-' || ? || ' days')
	`

	_, err := r.db.Exec(query, daysOld)
	return err
}

// DeleteOldMessagesByServer removes messages older than specified days for a server.
func (r *MessageRepository) DeleteOldMessagesByServer(serverID int, daysOld int) error {
	query := `
		DELETE FROM messages
		WHERE server_id = ? AND created_at < datetime('now', '-' || ? || ' days')
	`

	_, err := r.db.Exec(query, serverID, daysOld)
	return err
}

// CountMessagesByServer counts total messages for a server
func (r *MessageRepository) CountMessagesByServer(serverID int) (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE server_id = ?`

	var count int
	err := r.db.QueryRow(query, serverID).Scan(&count)
	return count, err
}
