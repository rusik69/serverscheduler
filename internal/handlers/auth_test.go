package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// SetupTestRouter creates a test router with all necessary routes
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Auth routes (public)
	router.POST("/api/auth/register", Register)
	router.POST("/api/auth/login", Login)

	// Protected routes
	protected := router.Group("/api/auth")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/user", GetUserInfo)
	}

	// Server routes
	router.GET("/api/servers", GetServers)
	router.POST("/api/servers", CreateServer)
	router.GET("/api/servers/:id", GetServer)
	router.PUT("/api/servers/:id", UpdateServer)
	router.DELETE("/api/servers/:id", DeleteServer)

	// Reservation routes
	router.GET("/api/reservations", GetReservations)
	router.POST("/api/reservations", CreateReservation)
	router.GET("/api/reservations/:id", GetReservation)

	return router
}

// GetAuthToken performs login and returns the auth token
func getAuthToken(t *testing.T, router *gin.Engine, username, password string) string {
	// Create login request
	loginData := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(loginData)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Extract token
	token, exists := response["token"]
	assert.True(t, exists)
	return token.(string)
}

func TestRegister(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := setupTestRouter()

	// Test cases
	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name: "Valid registration",
			payload: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid registration - missing username",
			payload: map[string]string{
				"password": "testpass",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestLogin(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	// First register a user through the API to ensure proper password hashing
	router := setupTestRouter()

	// Register user
	registerData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	jsonData, _ := json.Marshal(registerData)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Test cases
	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name: "Valid login",
			payload: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid login - wrong password",
			payload: map[string]string{
				"username": "testuser",
				"password": "wrongpass",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetUserInfo(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := setupTestRouter()

	// Register user first
	registerData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	jsonData, _ := json.Marshal(registerData)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get auth token by logging in
	token := getAuthToken(t, router, "testuser", "testpass")

	// Test cases
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/auth/user", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestChangePassword(t *testing.T) {
	// Setup test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	// Create a test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("oldpassword123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	_, err = database.GetDB().Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"testuser", string(hashedPassword), "user")
	assert.NoError(t, err)

	// Get the user ID
	var userID int64
	err = database.GetDB().QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userID         interface{}
		username       interface{}
		role           interface{}
		requestBody    map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "Valid password change",
			userID:   userID,
			username: "testuser",
			role:     "user",
			requestBody: map[string]string{
				"current_password": "oldpassword123",
				"new_password":     "newpassword456",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Invalid current password",
			userID:   userID,
			username: "testuser",
			role:     "user",
			requestBody: map[string]string{
				"current_password": "wrongpassword",
				"new_password":     "newpassword456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Current password is incorrect",
		},
		{
			name:     "Missing current password",
			userID:   userID,
			username: "testuser",
			role:     "user",
			requestBody: map[string]string{
				"new_password": "newpassword456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'ChangePasswordRequest.CurrentPassword' Error:Field validation for 'CurrentPassword' failed on the 'required' tag",
		},
		{
			name:     "Missing new password",
			userID:   userID,
			username: "testuser",
			role:     "user",
			requestBody: map[string]string{
				"current_password": "oldpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Key: 'ChangePasswordRequest.NewPassword' Error:Field validation for 'NewPassword' failed on the 'required' tag",
		},
		{
			name:     "New password too short",
			userID:   userID,
			username: "testuser",
			role:     "user",
			requestBody: map[string]string{
				"current_password": "oldpassword123",
				"new_password":     "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "New password must be at least 6 characters long",
		},
		{
			name:     "New password same as current",
			userID:   userID,
			username: "testuser",
			role:     "user",
			requestBody: map[string]string{
				"current_password": "oldpassword123",
				"new_password":     "oldpassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "New password must be different from current password",
		},
		{
			name: "No authentication context",
			requestBody: map[string]string{
				"current_password": "oldpassword123",
				"new_password":     "newpassword456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "User not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			jsonData, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/api/auth/change-password", bytes.NewBuffer(jsonData))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Set authentication context if provided
			if tt.userID != nil {
				c.Set("userID", tt.userID)
			}
			if tt.username != nil {
				c.Set("username", tt.username)
			}
			if tt.role != nil {
				c.Set("role", tt.role)
			}

			// Call handler
			ChangePassword(c)

			// Check response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				assert.Equal(t, tt.expectedError, response["error"])
			} else {
				assert.Equal(t, "Password changed successfully", response["message"])

				// Verify password was actually changed
				var newHashedPassword string
				err = database.GetDB().QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&newHashedPassword)
				assert.NoError(t, err)

				// Verify new password works
				err = bcrypt.CompareHashAndPassword([]byte(newHashedPassword), []byte("newpassword456"))
				assert.NoError(t, err)

				// Verify old password no longer works
				err = bcrypt.CompareHashAndPassword([]byte(newHashedPassword), []byte("oldpassword123"))
				assert.Error(t, err)
			}
		})
	}
}
