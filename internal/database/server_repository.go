package database

import (
	"database/sql"
	"log"

	"github.com/devalexandre/broker-ui/internal/messaging"
	"github.com/devalexandre/broker-ui/internal/models"
)

type ServerRepository struct {
	db *sql.DB
}

// NewServerRepository creates a new server repository
func NewServerRepository(db *sql.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

// Save saves a server to the database
func (r *ServerRepository) Save(name, url string, providerType messaging.ProviderType) error {
	if name == "" || url == "" {
		return nil
	}

	stmt, err := r.db.Prepare("INSERT INTO servers(name, url, provider_type) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, url, string(providerType))
	if err != nil {
		return err
	}

	log.Println("Server saved:", name, url, "Provider:", providerType)
	return nil
}

// Update updates a server in the database
func (r *ServerRepository) Update(serverID int, name, url string, providerType messaging.ProviderType) error {
	stmt, err := r.db.Prepare("UPDATE servers SET name = ?, url = ?, provider_type = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, url, string(providerType), serverID)
	if err != nil {
		return err
	}

	log.Println("Server updated:", name, url, "Provider:", providerType)
	return nil
}

// GetAll loads all servers from the database
func (r *ServerRepository) GetAll() ([]models.Server, error) {
	rows, err := r.db.Query("SELECT id, name, url, COALESCE(provider_type, 'NATS') FROM servers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var s models.Server
		var providerTypeStr string
		err := rows.Scan(&s.ID, &s.Name, &s.URL, &providerTypeStr)
		if err != nil {
			return nil, err
		}
		s.ProviderType = messaging.ProviderType(providerTypeStr)
		servers = append(servers, s)
	}

	return servers, nil
}
