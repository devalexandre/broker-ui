package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db      *sql.DB
	Servers []Server
	Topics  []Topic
	Subs    []Sub
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "./servers.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, url TEXT, ctype TEXT)`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS topics (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, topic_name TEXT)`)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS subs (id INTEGER PRIMARY KEY AUTOINCREMENT, server_id INTEGER, sub_name TEXT)`)
	if err != nil {
		return nil, err
	}

	// Create settings table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT)`)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d Database) LoadServers() []Server {
	rows, err := d.db.Query("SELECT id, name, url, ctype FROM servers")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	d.Servers = []Server{}
	for rows.Next() {
		var s Server
		err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Client)
		if err != nil {
			log.Fatal(err)
		}
		d.Servers = append(d.Servers, s)
	}

	return d.Servers
}

func (d Database) SaveServer(name, url, ctype string) {
	stmt, err := d.db.Prepare("INSERT INTO servers(name, url,ctype) VALUES(?, ?,?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, url, ctype)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server saved:", name, url, ctype)
}

func (d Database) UpdateServer(serverID int, name, url string) {
	stmt, err := d.db.Prepare("UPDATE servers SET name = ?, url = ? WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, url, serverID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server updated:", name, url)
}

func (d Database) LoadTopics(serverID int) []Topic {

	rows, err := d.db.Query("SELECT id, topic_name FROM topics WHERE server_id = ?", serverID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var topics = []Topic{}
	for rows.Next() {
		var t Topic
		err := rows.Scan(&t.ID, &t.TopicName)
		if err != nil {
			log.Fatal(err)
		}
		topics = append(topics, t)
	}

	log.Printf("Loaded %v topics for server %v", len(topics), serverID)
	return topics
}

func (d Database) SaveTopic(serverID int, topicName string) {
	if topicName == "" {
		return
	}

	stmt, err := d.db.Prepare("INSERT INTO topics(server_id, topic_name) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID, topicName)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Topic saved:", topicName)
}

func (d Database) LoadSubs(serverID int) []Sub {
	rows, err := d.db.Query("SELECT id, sub_name FROM subs WHERE server_id = ?", serverID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	subs := []Sub{}
	for rows.Next() {
		var s Sub
		err := rows.Scan(&s.ID, &s.SubName)
		if err != nil {
			log.Fatal(err)
		}
		subs = append(subs, s)
	}

	return subs
}

func (d Database) SaveSub(serverID int, subName string) {
	if subName == "" {
		return
	}

	stmt, err := d.db.Prepare("INSERT INTO subs(server_id, sub_name) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID, subName)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Sub saved:", subName)
}

// DeleteServer
func (d Database) DeleteServer(serverID int) error {
	stmt, err := d.db.Prepare("DELETE FROM servers WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverID)
	if err != nil {
		return err
	}

	log.Println("Server deleted:", serverID)

	return nil
}

func (d *Database) GetSetting(key string) (string, error) {
	var value string
	err := d.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil // No setting found, return empty string
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (d *Database) SetSetting(key string, value string) error {
	// Use INSERT OR REPLACE to update existing setting or insert new one
	stmt, err := d.db.Prepare("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(key, value)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTopic deletes a topic from the database
func (d *Database) DeleteTopic(serverId int, topicName string) error {
	stmt, err := d.db.Prepare("DELETE FROM topics WHERE server_id = ? AND topic_name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverId, topicName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Println("Topic deleted:", topicName)
	return nil
}

// DeleteSub deletes a subscription from the database
func (d *Database) DeleteSub(serverId int, subName string) error {
	stmt, err := d.db.Prepare("DELETE FROM subs WHERE server_id = ? AND sub_name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverId, subName)
	if err != nil {
		fmt.Println(err)
		return err
	}

	log.Println("Sub deleted:", subName)
	return nil
}
