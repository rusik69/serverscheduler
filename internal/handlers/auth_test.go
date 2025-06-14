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
