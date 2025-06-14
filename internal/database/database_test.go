package database

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rusik69/serverscheduler/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) {
	err := InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

func teardownTestDB() {
	CleanupTestDB()
}

func TestInitDB(t *testing.T) {
	// Clean up any existing test database
	os.Remove("test.db")
	defer os.Remove("test.db")

	// Initialize database
	err := InitTestDB()
	assert.NoError(t, err)

	// Check if database connection exists
	db := GetDB()
	assert.NotNil(t, db)

	// Check if tables exist by querying them
	_, err = db.Exec("SELECT COUNT(*) FROM users")
	assert.NoError(t, err, "Users table should exist")

	_, err = db.Exec("SELECT COUNT(*) FROM servers")
	assert.NoError(t, err, "Servers table should exist")

	_, err = db.Exec("SELECT COUNT(*) FROM reservations")
	assert.NoError(t, err, "Reservations table should exist")

	CleanupTestDB()
}

func TestInitTestDB(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)
	defer CleanupTestDB()

	db := GetDB()
	assert.NotNil(t, db)
}

func TestCleanupTestDB(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)

	// Insert test data
	_, err = GetDB().Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "testuser", "testpass", "user")
	assert.NoError(t, err)

	// Cleanup
	CleanupTestDB()

	// Database should be closed (but the pointer might still exist)
	// Just verify we can't execute queries anymore
	db := GetDB()
	if db != nil {
		_, err = db.Exec("SELECT COUNT(*) FROM users")
		// Should get an error because database is closed
		assert.Error(t, err)
	}
}

func TestCloseDB(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)

	Close()
	// Database should be closed (but the pointer might still exist)
	// Just verify we can't execute queries anymore
	db := GetDB()
	if db != nil {
		_, err = db.Exec("SELECT COUNT(*) FROM users")
		// Should get an error because database is closed
		assert.Error(t, err)
	}
}

func TestUserCRUD(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)
	defer CleanupTestDB()

	db := GetDB()

	// Create user
	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "testuser", "testpass", "user")
	assert.NoError(t, err)

	// Read user
	var user models.User
	err = db.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", "testuser").Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "user", user.Role)

	// Update user
	_, err = db.Exec("UPDATE users SET role = ? WHERE username = ?", "admin", "testuser")
	assert.NoError(t, err)

	// Verify update
	err = db.QueryRow("SELECT role FROM users WHERE username = ?", "testuser").Scan(&user.Role)
	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Role)

	// Delete user
	_, err = db.Exec("DELETE FROM users WHERE username = ?", "testuser")
	assert.NoError(t, err)

	// Verify deletion
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&user.ID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestServerCRUD(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)
	defer CleanupTestDB()

	db := GetDB()

	// Create server
	_, err = db.Exec("INSERT INTO servers (name, status) VALUES (?, ?)", "test-server", "available")
	assert.NoError(t, err)

	// Read server
	var server models.Server
	err = db.QueryRow("SELECT id, name, status FROM servers WHERE name = ?", "test-server").Scan(&server.ID, &server.Name, &server.Status)
	assert.NoError(t, err)
	assert.Equal(t, "test-server", server.Name)
	assert.Equal(t, "available", server.Status)

	// Update server
	_, err = db.Exec("UPDATE servers SET status = ? WHERE name = ?", "reserved", "test-server")
	assert.NoError(t, err)

	// Verify update
	err = db.QueryRow("SELECT status FROM servers WHERE name = ?", "test-server").Scan(&server.Status)
	assert.NoError(t, err)
	assert.Equal(t, "reserved", server.Status)

	// Delete server
	_, err = db.Exec("DELETE FROM servers WHERE name = ?", "test-server")
	assert.NoError(t, err)

	// Verify deletion
	err = db.QueryRow("SELECT id FROM servers WHERE name = ?", "test-server").Scan(&server.ID)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestReservationCRUD(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)
	defer CleanupTestDB()

	db := GetDB()

	// Create test user and server first
	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "testuser", "testpass", "user")
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO servers (name, status) VALUES (?, ?)", "test-server", "available")
	assert.NoError(t, err)

	// Get user and server IDs
	var userID, serverID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)
	assert.NoError(t, err)

	err = db.QueryRow("SELECT id FROM servers WHERE name = ?", "test-server").Scan(&serverID)
	assert.NoError(t, err)

	// Create reservation
	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	_, err = db.Exec("INSERT INTO reservations (user_id, server_id, server_name, start_time, end_time, status) VALUES (?, ?, ?, ?, ?, ?)",
		userID, serverID, "test-server", startTime, endTime, "active")
	assert.NoError(t, err)

	// Read reservation
	var reservation models.Reservation
	err = db.QueryRow("SELECT id, user_id, server_id, server_name, start_time, end_time, status FROM reservations WHERE user_id = ?", userID).
		Scan(&reservation.ID, &reservation.UserID, &reservation.ServerID, &reservation.ServerName, &reservation.StartTime, &reservation.EndTime, &reservation.Status)
	assert.NoError(t, err)
	assert.Equal(t, int64(userID), reservation.UserID)
	assert.Equal(t, int64(serverID), reservation.ServerID)
	assert.Equal(t, "test-server", reservation.ServerName)
	assert.Equal(t, "active", reservation.Status)
}

func TestCreateUser(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)
	defer CleanupTestDB()

	db := GetDB()

	// Test creating a user
	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "newuser", "newpass", "user")
	assert.NoError(t, err)

	// Verify user was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "newuser").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreateServer(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)
	defer CleanupTestDB()

	db := GetDB()

	// Test creating a server
	_, err = db.Exec("INSERT INTO servers (name, status) VALUES (?, ?)", "srv1", "available")
	assert.NoError(t, err)

	// Verify server was created
	var name, status string
	err = db.QueryRow("SELECT name, status FROM servers WHERE name = ?", "srv1").Scan(&name, &status)
	assert.NoError(t, err)
	assert.Equal(t, "srv1", name)
	assert.Equal(t, "available", status)
}

func TestCreateReservation(t *testing.T) {
	err := InitTestDB()
	assert.NoError(t, err)
	defer CleanupTestDB()

	db := GetDB()

	// Create test user and server first
	_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "testuser", "testpass", "user")
	assert.NoError(t, err)

	_, err = db.Exec("INSERT INTO servers (name, status) VALUES (?, ?)", "test-server", "available")
	assert.NoError(t, err)

	// Get IDs
	var userID, serverID int
	err = db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)
	assert.NoError(t, err)

	err = db.QueryRow("SELECT id FROM servers WHERE name = ?", "test-server").Scan(&serverID)
	assert.NoError(t, err)

	// Create reservation
	startTime := time.Now()
	endTime := startTime.Add(time.Hour)
	_, err = db.Exec("INSERT INTO reservations (user_id, server_id, server_name, start_time, end_time, status) VALUES (?, ?, ?, ?, ?, ?)",
		userID, serverID, "test-server", startTime, endTime, "active")
	assert.NoError(t, err)

	// Verify reservation was created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM reservations WHERE user_id = ?", userID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}
