package database

import (
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rusik69/serverscheduler/internal/models"
)

func setupTestDB(t *testing.T) {
	err := InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

func teardownTestDB() {
	CleanupTestDB()
	DB = nil // Set DB to nil after closing
}

func TestInitDB(t *testing.T) {
	err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()

	// Check if tables exist
	_, err = DB.Query("SELECT * FROM users LIMIT 1")
	if err != nil {
		t.Fatalf("Failed to query users table: %v", err)
	}
}

func TestInitTestDB(t *testing.T) {
	err := InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer teardownTestDB()

	// Check if tables exist
	_, err = DB.Query("SELECT * FROM users LIMIT 1")
	if err != nil {
		t.Fatalf("Failed to query users table: %v", err)
	}
}

func TestCleanupTestDB(t *testing.T) {
	err := InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Insert test data
	_, err = DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "testuser", "testpass", "user")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Clean up
	err = CleanupTestDB()
	if err != nil {
		t.Fatalf("Failed to cleanup test database: %v", err)
	}
	DB = nil // Set DB to nil after closing

	// Verify tables are dropped
	err = InitTestDB()
	if err != nil {
		t.Fatalf("Failed to reinitialize test database: %v", err)
	}
	defer teardownTestDB()

	// Try to query the users table - should be empty
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query users table: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 users after cleanup, got %d", count)
	}
}

func TestCloseDB(t *testing.T) {
	err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	err = CloseDB()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}
	DB = nil // Set DB to nil after closing

	if DB != nil {
		t.Error("Expected DB to be nil after closing")
	}
}

func TestUserCRUD(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB()

	// Create user
	_, err := DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "alice", "alicepass", "user")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Retrieve user
	row := DB.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", "alice")
	var user models.User
	var id int64
	if err := row.Scan(&id, &user.Username, &user.Password, &user.Role); err != nil {
		t.Fatalf("Failed to scan user: %v", err)
	}
	if user.Username != "alice" || user.Role != "user" {
		t.Errorf("Unexpected user data: %+v", user)
	}
}

func TestServerCRUD(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB()

	// Create server
	_, err := DB.Exec("INSERT INTO servers (name, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		"srv1", "desc1", "available", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert server: %v", err)
	}

	// Retrieve server
	row := DB.QueryRow("SELECT id, name, description, status FROM servers WHERE name = ?", "srv1")
	var id int64
	var name, desc, status string
	if err := row.Scan(&id, &name, &desc, &status); err != nil {
		t.Fatalf("Failed to scan server: %v", err)
	}
	if name != "srv1" || status != "available" {
		t.Errorf("Unexpected server data: name=%s, status=%s", name, status)
	}
}

func TestReservationCRUD(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB()

	// Insert user and server
	res, err := DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "bob", "bobpass", "user")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	userID, _ := res.LastInsertId()
	res, err = DB.Exec("INSERT INTO servers (name, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		"srv2", "desc2", "available", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert server: %v", err)
	}
	serverID, _ := res.LastInsertId()

	// Create reservation
	start := time.Now()
	end := start.Add(2 * time.Hour)
	_, err = DB.Exec("INSERT INTO reservations (server_id, user_id, start_time, end_time) VALUES (?, ?, ?, ?)",
		serverID, userID, start, end)
	if err != nil {
		t.Fatalf("Failed to insert reservation: %v", err)
	}

	// Retrieve reservation
	row := DB.QueryRow("SELECT server_id, user_id FROM reservations WHERE user_id = ?", userID)
	var gotServerID, gotUserID int64
	if err := row.Scan(&gotServerID, &gotUserID); err != nil {
		t.Fatalf("Failed to scan reservation: %v", err)
	}
	if gotServerID != serverID || gotUserID != userID {
		t.Errorf("Unexpected reservation data: server_id=%d, user_id=%d", gotServerID, gotUserID)
	}
}
