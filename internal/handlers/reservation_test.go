package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupReservationTestDB(t *testing.T) {
	err := database.InitTestDB()
	assert.NoError(t, err)

	// Insert test data
	_, err = database.GetDB().Exec("INSERT INTO users (id, username, password, role) VALUES (1, 'testuser', 'password', 'user')")
	assert.NoError(t, err)
	_, err = database.GetDB().Exec("INSERT INTO users (id, username, password, role) VALUES (2, 'root', 'password', 'root')")
	assert.NoError(t, err)

	_, err = database.GetDB().Exec("INSERT INTO servers (id, name, status) VALUES (1, 'test-server', 'available')")
	assert.NoError(t, err)
	_, err = database.GetDB().Exec("INSERT INTO servers (id, name, status) VALUES (2, 'reserved-server', 'reserved')")
	assert.NoError(t, err)
}

func TestCreateReservation(t *testing.T) {
	setupReservationTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		username       string
		role           string
		reservation    models.Reservation
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "Valid reservation",
			userID:   1,
			username: "testuser",
			role:     "user",
			reservation: models.Reservation{
				ServerID:  1,
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "Start time after end time",
			userID:   1,
			username: "testuser",
			role:     "user",
			reservation: models.Reservation{
				ServerID:  1,
				StartTime: time.Now().Add(2 * time.Hour),
				EndTime:   time.Now().Add(1 * time.Hour),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Start time must be before end time",
		},
		{
			name:     "Start time in the past",
			userID:   1,
			username: "testuser",
			role:     "user",
			reservation: models.Reservation{
				ServerID:  1,
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now().Add(1 * time.Hour),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Start time cannot be in the past",
		},
		{
			name:     "Server not found",
			userID:   1,
			username: "testuser",
			role:     "user",
			reservation: models.Reservation{
				ServerID:  999,
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Server not found",
		},
		{
			name:     "Server not available",
			userID:   1,
			username: "testuser",
			role:     "user",
			reservation: models.Reservation{
				ServerID:  2,
				StartTime: time.Now().Add(1 * time.Hour),
				EndTime:   time.Now().Add(2 * time.Hour),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Server is not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set context values
			c.Set("userID", tt.userID)
			c.Set("username", tt.username)
			c.Set("role", tt.role)

			// Create request body
			body, _ := json.Marshal(tt.reservation)
			c.Request = httptest.NewRequest("POST", "/api/reservations", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			CreateReservation(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestGetReservations(t *testing.T) {
	setupReservationTestDB(t)
	gin.SetMode(gin.TestMode)

	// Create test reservations
	_, err := database.GetDB().Exec(`
		INSERT INTO reservations (server_id, user_id, server_name, start_time, end_time, status) 
		VALUES (1, 1, 'test-server', ?, ?, 'active')`,
		time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour))
	assert.NoError(t, err)

	_, err = database.GetDB().Exec(`
		INSERT INTO reservations (server_id, user_id, server_name, start_time, end_time, status) 
		VALUES (1, 2, 'test-server', ?, ?, 'active')`,
		time.Now().Add(3*time.Hour), time.Now().Add(4*time.Hour))
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userID         int64
		role           string
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "Regular user sees only own reservations",
			userID:         1,
			role:           "user",
			expectedCount:  1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Root user sees all reservations",
			userID:         2,
			role:           "root",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)
			c.Request = httptest.NewRequest("GET", "/api/reservations", nil)

			GetReservations(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var reservations []models.Reservation
			json.Unmarshal(w.Body.Bytes(), &reservations)
			assert.Equal(t, tt.expectedCount, len(reservations))
		})
	}
}

func TestGetReservation(t *testing.T) {
	setupReservationTestDB(t)
	gin.SetMode(gin.TestMode)

	// Create test reservation
	result, err := database.GetDB().Exec(`
		INSERT INTO reservations (server_id, user_id, server_name, start_time, end_time, status) 
		VALUES (1, 1, 'test-server', ?, ?, 'active')`,
		time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour))
	assert.NoError(t, err)
	reservationID, _ := result.LastInsertId()

	tests := []struct {
		name           string
		userID         int64
		role           string
		reservationID  string
		expectedStatus int
		shouldFind     bool
	}{
		{
			name:           "User can access own reservation",
			userID:         1,
			role:           "user",
			reservationID:  strconv.FormatInt(reservationID, 10),
			expectedStatus: http.StatusOK,
			shouldFind:     true,
		},
		{
			name:           "User cannot access other's reservation",
			userID:         2,
			role:           "user",
			reservationID:  strconv.FormatInt(reservationID, 10),
			expectedStatus: http.StatusNotFound,
			shouldFind:     false,
		},
		{
			name:           "Root can access any reservation",
			userID:         2,
			role:           "root",
			reservationID:  strconv.FormatInt(reservationID, 10),
			expectedStatus: http.StatusOK,
			shouldFind:     true,
		},
		{
			name:           "Invalid reservation ID",
			userID:         1,
			role:           "user",
			reservationID:  "invalid",
			expectedStatus: http.StatusBadRequest,
			shouldFind:     false,
		},
		{
			name:           "Reservation not found",
			userID:         1,
			role:           "user",
			reservationID:  "999",
			expectedStatus: http.StatusNotFound,
			shouldFind:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)
			c.Params = gin.Params{{Key: "id", Value: tt.reservationID}}
			c.Request = httptest.NewRequest("GET", "/api/reservations/"+tt.reservationID, nil)

			GetReservation(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCancelReservation(t *testing.T) {
	setupReservationTestDB(t)
	gin.SetMode(gin.TestMode)

	// Create test reservation
	result, err := database.GetDB().Exec(`
		INSERT INTO reservations (server_id, user_id, server_name, start_time, end_time, status) 
		VALUES (1, 1, 'test-server', ?, ?, 'active')`,
		time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour))
	assert.NoError(t, err)
	reservationID, _ := result.LastInsertId()

	tests := []struct {
		name           string
		userID         int64
		role           string
		reservationID  string
		expectedStatus int
	}{
		{
			name:           "User can cancel own reservation",
			userID:         1,
			role:           "user",
			reservationID:  strconv.FormatInt(reservationID, 10),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid reservation ID",
			userID:         1,
			role:           "user",
			reservationID:  "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)
			c.Params = gin.Params{{Key: "id", Value: tt.reservationID}}
			c.Request = httptest.NewRequest("DELETE", "/api/reservations/"+tt.reservationID, nil)

			CancelReservation(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCleanupExpiredReservations(t *testing.T) {
	setupReservationTestDB(t)

	// Create expired reservation
	_, err := database.GetDB().Exec(`
		INSERT INTO reservations (server_id, user_id, server_name, start_time, end_time, status) 
		VALUES (1, 1, 'test-server', ?, ?, 'active')`,
		time.Now().Add(-2*time.Hour), time.Now().Add(-1*time.Hour))
	assert.NoError(t, err)

	CleanupExpiredReservations()

	// Check that reservation is marked as expired
	var status string
	err = database.GetDB().QueryRow("SELECT status FROM reservations WHERE user_id = 1").Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, "expired", status)
}
