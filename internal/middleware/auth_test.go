package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
)

func TestAuthMiddleware(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	// Create a test user
	user := models.User{
		Username: "testuser",
		Password: "testpass",
		Role:     "user",
	}
	_, err = database.DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		user.Username, user.Password, user.Role)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Generate token
	token, err := GenerateToken(1, user.Username, user.Role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Set up test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

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
		{
			name:           "No token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check response
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAdminMiddleware(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	// Create test users
	adminUser := models.User{
		Username: "admin",
		Password: "adminpass",
		Role:     "admin",
	}
	regularUser := models.User{
		Username: "user",
		Password: "userpass",
		Role:     "user",
	}

	// Insert users
	_, err = database.DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		adminUser.Username, adminUser.Password, adminUser.Role)
	if err != nil {
		t.Fatalf("Failed to insert admin user: %v", err)
	}
	_, err = database.DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		regularUser.Username, regularUser.Password, regularUser.Role)
	if err != nil {
		t.Fatalf("Failed to insert regular user: %v", err)
	}

	// Generate tokens
	adminToken, err := GenerateToken(1, adminUser.Username, adminUser.Role)
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}
	userToken, err := GenerateToken(2, regularUser.Username, regularUser.Role)
	if err != nil {
		t.Fatalf("Failed to generate user token: %v", err)
	}

	// Set up test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware(), AdminMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test cases
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Admin token",
			token:          adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Regular user token",
			token:          userToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "No token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			// Perform request
			router.ServeHTTP(w, req)

			// Check response
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
