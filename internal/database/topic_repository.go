package database

import (
	"database/sql"
	"log"

	"github.com/devalexandre/broker-ui/internal/models"
)

type TopicRepository struct {
	db *sql.DB
}

// NewTopicRepository creates a new topic repository
func NewTopicRepository(db *sql.DB) *TopicRepository {
	return &TopicRepository{db: db}
}

// Save saves a topic to the database
func (r *TopicRepository) Save(serverID int, topicName string) error {
	if topicName == "" {
		return nil
	}

	stmt, err := r.db.Prepare("INSERT INTO topics(server_id, topic_name) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID, topicName)
	if err != nil {
		return err
	}

	log.Println("Topic saved:", topicName)
	return nil
}

// GetByServerID loads topics for a specific server
func (r *TopicRepository) GetByServerID(serverID int) ([]models.Topic, error) {
	rows, err := r.db.Query("SELECT id, topic_name FROM topics WHERE server_id = ?", serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		err := rows.Scan(&t.ID, &t.TopicName)
		if err != nil {
			return nil, err
		}
		t.ServerID = serverID
		topics = append(topics, t)
	}

	return topics, nil
}

// Delete deletes a topic from the database
func (r *TopicRepository) Delete(topicName string, serverID int) error {
	stmt, err := r.db.Prepare("DELETE FROM topics WHERE topic_name = ? AND server_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(topicName, serverID)
	if err != nil {
		return err
	}

	log.Println("Topic deleted:", topicName)
	return nil
}
