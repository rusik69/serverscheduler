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
	"github.com/rusik69/serverscheduler/internal/models"
	"github.com/stretchr/testify/assert"
)

// setupServerTestRouter creates a test router with server routes
func setupServerTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Auth routes
	router.POST("/api/auth/register", Register)
	router.POST("/api/auth/login", Login)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/servers", GetServers)
		protected.POST("/servers", middleware.AdminMiddleware(), CreateServer)
		protected.GET("/servers/:id", GetServer)
	}

	return router
}

// getServerAuthToken performs login and returns the auth token
func getServerAuthToken(t *testing.T, router *gin.Engine, username, password string) string {
	loginData := map[string]string{
		"username": username,
		"password": password,
	}
	jsonData, _ := json.Marshal(loginData)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token, exists := response["token"]
	if !exists {
		t.Fatal("Login response missing token")
	}
	return token.(string)
}

func TestCreateServer(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := setupServerTestRouter()

	// Register test users first
	// Register root user
	registerData := map[string]string{
		"username": "root",
		"password": "rootpass",
	}
	jsonData, _ := json.Marshal(registerData)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Update user role to admin in database
	_, err = database.GetDB().Exec("UPDATE users SET role = ? WHERE username = ?", "admin", "root")
	if err != nil {
		t.Fatalf("Failed to update root user role: %v", err)
	}

	// Register regular user
	registerData = map[string]string{
		"username": "user",
		"password": "userpass",
	}
	jsonData, _ = json.Marshal(registerData)
	req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get tokens
	rootToken := getServerAuthToken(t, router, "root", "rootpass")
	userToken := getServerAuthToken(t, router, "user", "userpass")

	tests := []struct {
		name           string
		token          string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name:  "Create server as root",
			token: rootToken,
			payload: map[string]string{
				"name": "test-server",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:  "Create server as regular user",
			token: userToken,
			payload: map[string]string{
				"name": "test-server-2",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:  "Create server without auth",
			token: "",
			payload: map[string]string{
				"name": "test-server-3",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestListServers(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	// Register test user
	router := setupServerTestRouter()
	registerData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	jsonData, _ := json.Marshal(registerData)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Create test server
	_, err = database.GetDB().Exec(
		"INSERT INTO servers (name) VALUES (?)",
		"test-server",
	)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	token := getServerAuthToken(t, router, "testuser", "testpass")

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "List servers with auth",
			token:          token,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "List servers without auth",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/servers", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string][]models.Server
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "servers")
			}
		})
	}
}
