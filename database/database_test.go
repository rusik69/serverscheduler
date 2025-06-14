package database

import (
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rusik69/serverscheduler/models"
)

func TestInitDB(t *testing.T) {
	// Test initialization
	err := InitDB()
	if err != nil {
		t.Errorf("Failed to initialize database: %v", err)
	}

	// Test that tables are created
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query users table: %v", err)
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM servers").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query servers table: %v", err)
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM reservations").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query reservations table: %v", err)
	}
}

func TestInitTestDB(t *testing.T) {
	// Test test database initialization
	err := InitTestDB()
	if err != nil {
		t.Errorf("Failed to initialize test database: %v", err)
	}

	// Test that tables are created
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query users table: %v", err)
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM servers").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query servers table: %v", err)
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM reservations").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query reservations table: %v", err)
	}
}

func TestCleanupTestDB(t *testing.T) {
	// Initialize test database
	err := InitTestDB()
	if err != nil {
		t.Errorf("Failed to initialize test database: %v", err)
	}

	// Add some test data
	user := models.User{
		Username: "testuser",
		Password: "testpass",
		Role:     "user",
	}
	_, err = DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		user.Username, user.Password, user.Role)
	if err != nil {
		t.Errorf("Failed to insert test user: %v", err)
	}

	// Test cleanup
	err = CleanupTestDB()
	if err != nil {
		t.Errorf("Failed to cleanup test database: %v", err)
	}

	// Verify cleanup
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Errorf("Failed to query users table: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 users after cleanup, got %d", count)
	}
}

func TestCloseDB(t *testing.T) {
	// Initialize database
	err := InitDB()
	if err != nil {
		t.Errorf("Failed to initialize database: %v", err)
	}

	// Test closing
	err = CloseDB()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Verify that DB is closed
	if DB != nil {
		t.Error("Expected DB to be nil after closing")
	}
}

func setupTestDB(t *testing.T) {
	err := InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

func teardownTestDB() {
	CleanupTestDB()
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
	_, err = DB.Exec("INSERT INTO reservations (server_id, user_id, start_time, end_time, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		serverID, userID, start, end, "active", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert reservation: %v", err)
	}

	// Retrieve reservation
	row := DB.QueryRow("SELECT server_id, user_id, status FROM reservations WHERE user_id = ?", userID)
	var gotServerID, gotUserID int64
	var status string
	if err := row.Scan(&gotServerID, &gotUserID, &status); err != nil {
		t.Fatalf("Failed to scan reservation: %v", err)
	}
	if gotServerID != serverID || gotUserID != userID || status != "active" {
		t.Errorf("Unexpected reservation data: server_id=%d, user_id=%d, status=%s", gotServerID, gotUserID, status)
	}
}
