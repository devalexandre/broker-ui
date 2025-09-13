package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

// New creates a new database instance and initializes tables
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// GetDB returns the database connection
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// createTables creates all necessary tables
func (d *Database) createTables() error {
	// Create servers table
	_, err := d.db.Exec(`CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, url TEXT, provider_type TEXT DEFAULT 'NATS')`)
	if err != nil {
		return err
	}

	// Migration: add provider_type column if it doesn't exist
	_, err = d.db.Exec(`ALTER TABLE servers ADD COLUMN provider_type TEXT DEFAULT 'NATS'`)
	if err != nil {
		// Ignore error if column already exists
		log.Printf("Column provider_type may already exist: %v", err)
	}

	// Create topics table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS topics (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, topic_name TEXT)`)
	if err != nil {
		return err
	}

	// Create subs table
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS subs (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, sub_name TEXT, subject_pattern TEXT)`)
	if err != nil {
		return err
	}

	// Migration: add subject_pattern column if it doesn't exist
	_, err = d.db.Exec(`ALTER TABLE subs ADD COLUMN subject_pattern TEXT DEFAULT ''`)
	if err != nil {
		// Ignore error if column already exists
		log.Printf("Column subject_pattern may already exist: %v", err)
	}

	return nil
}
