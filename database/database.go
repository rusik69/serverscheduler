package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./server_scheduler.db"
	}
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Create tables if they don't exist
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS reservations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			FOREIGN KEY (server_id) REFERENCES servers(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	return nil
}

// InitTestDB initializes a test database with proper schema
func InitTestDB() error {
	// Use in-memory SQLite for tests
	var err error
	DB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return fmt.Errorf("failed to open test database: %v", err)
	}

	// Create tables
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS reservations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			FOREIGN KEY (server_id) REFERENCES servers(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create test tables: %v", err)
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// CleanupTestDB cleans up the test database
func CleanupTestDB() error {
	if DB != nil {
		// Drop all tables
		_, err := DB.Exec(`
			DROP TABLE IF EXISTS reservations;
			DROP TABLE IF EXISTS servers;
			DROP TABLE IF EXISTS users;
		`)
		if err != nil {
			return fmt.Errorf("failed to drop test tables: %v", err)
		}
		return DB.Close()
	}
	return nil
}

// Cleanup removes the database file
func Cleanup() {
	os.Remove("./serverscheduler.db")
}
