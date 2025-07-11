package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Init initializes the database connection and creates tables
func Init() error {
	var err error
	db, err = sql.Open("sqlite3", "./data/serverscheduler.db")
	if err != nil {
		return err
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user'
		)
	`)
	if err != nil {
		return err
	}

	// Create servers table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL DEFAULT 'available',
			ip_address TEXT,
			username TEXT,
			password TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create reservations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reservations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			server_id INTEGER NOT NULL,
			server_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (server_id) REFERENCES servers(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// CreateRootUser creates a root user if it doesn't exist
func CreateRootUser() error {
	// Check if root user exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'root'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create root user
		_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
			"root",
			"root", // In production, this should be a hashed password
			"admin")
		if err != nil {
			return err
		}
		log.Println("Root user created")
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// InitDB initializes the database connection
func InitDB() error {
	var err error
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./server_scheduler.db"
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
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

// InitTestDB initializes a test database
func InitTestDB() error {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user'
		)
	`)
	if err != nil {
		return err
	}

	// Create servers table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS servers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL DEFAULT 'available',
			ip_address TEXT,
			username TEXT,
			password TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create reservations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reservations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			server_id INTEGER NOT NULL,
			server_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (server_id) REFERENCES servers(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// CleanupTestDB cleans up the test database
func CleanupTestDB() error {
	if db != nil {
		// Drop all tables
		_, err := db.Exec(`
			DROP TABLE IF EXISTS reservations;
			DROP TABLE IF EXISTS servers;
			DROP TABLE IF EXISTS users;
		`)
		if err != nil {
			return fmt.Errorf("failed to drop test tables: %v", err)
		}
		return db.Close()
	}
	return nil
}

// Cleanup removes the database file
func Cleanup() {
	os.Remove("./serverscheduler.db")
}

// isColumnExistsError checks if the error is about a column already existing
func isColumnExistsError(err error) bool {
	return err != nil && (err.Error() == "duplicate column name: ip_address" ||
		err.Error() == "duplicate column name: username" ||
		err.Error() == "duplicate column name: password")
}

// RunMigrations runs database migrations for existing tables
func RunMigrations() error {
	// Add new columns to existing servers table if they don't exist
	_, err := db.Exec(`ALTER TABLE servers ADD COLUMN ip_address TEXT`)
	if err != nil && !isColumnExistsError(err) {
		return err
	}

	_, err = db.Exec(`ALTER TABLE servers ADD COLUMN username TEXT`)
	if err != nil && !isColumnExistsError(err) {
		return err
	}

	_, err = db.Exec(`ALTER TABLE servers ADD COLUMN password TEXT`)
	if err != nil && !isColumnExistsError(err) {
		return err
	}

	return nil
}
