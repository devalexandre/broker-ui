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
	if ok, checkErr := columnExists(d.db, "servers", "provider_type"); checkErr != nil {
		log.Printf("Failed to check provider_type column: %v", checkErr)
	} else if !ok {
		if _, err := d.db.Exec(`ALTER TABLE servers ADD COLUMN provider_type TEXT DEFAULT 'NATS'`); err != nil {
			return err
		}
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
	if ok, checkErr := columnExists(d.db, "subs", "subject_pattern"); checkErr != nil {
		log.Printf("Failed to check subject_pattern column: %v", checkErr)
	} else if !ok {
		if _, err := d.db.Exec(`ALTER TABLE subs ADD COLUMN subject_pattern TEXT DEFAULT ''`); err != nil {
			return err
		}
	}

	// Create messages table for audit purposes
	_, err = d.db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER,
		subject TEXT NOT NULL,
		payload TEXT,
		provider_type TEXT NOT NULL,
		direction TEXT NOT NULL,
		timestamp INTEGER,
		created_at TEXT NOT NULL,
		FOREIGN KEY (server_id) REFERENCES servers (id) ON DELETE CASCADE
	)`)
	if err != nil {
		return err
	}

	// Create index for better query performance
	_, err = d.db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_server_id ON messages (server_id)`)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at)`)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_subject ON messages (subject)`)
	if err != nil {
		return err
	}

	return nil
}

func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			colType    string
			notNull    int
			defaultV   sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultV, &primaryKey); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}
	return false, rows.Err()
}
