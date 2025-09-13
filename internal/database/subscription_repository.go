package database

import (
	"database/sql"
	"log"

	"github.com/devalexandre/broker-ui/internal/models"
)

type SubscriptionRepository struct {
	db *sql.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Save saves a subscription to the database
func (r *SubscriptionRepository) Save(serverID int, subName, subjectPattern string) error {
	if subName == "" || subjectPattern == "" {
		return nil
	}

	stmt, err := r.db.Prepare("INSERT INTO subs(server_id, sub_name, subject_pattern) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID, subName, subjectPattern)
	if err != nil {
		return err
	}

	log.Printf("Sub saved: %s (pattern: %s)", subName, subjectPattern)
	return nil
}

// GetByServerID loads subscriptions for a specific server
func (r *SubscriptionRepository) GetByServerID(serverID int) ([]models.Subscription, error) {
	rows, err := r.db.Query("SELECT id, sub_name, COALESCE(subject_pattern, '') FROM subs WHERE server_id = ?", serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var s models.Subscription
		err := rows.Scan(&s.ID, &s.SubName, &s.SubjectPattern)
		if err != nil {
			return nil, err
		}
		s.ServerID = serverID
		subs = append(subs, s)
	}

	return subs, nil
}

// Delete deletes a subscription from the database
func (r *SubscriptionRepository) Delete(subName string, serverID int) error {
	stmt, err := r.db.Prepare("DELETE FROM subs WHERE sub_name = ? AND server_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(subName, serverID)
	if err != nil {
		return err
	}

	log.Println("Sub deleted:", subName)
	return nil
}
