package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthMiddleware(t *testing.T) {
	// Initialize test database
	err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	// Hash password for test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Create test user
	_, err = database.GetDB().Exec(
		"INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		1, "testuser", string(hashedPassword), "user",
	)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate valid JWT token
	validToken, err := GenerateToken(1, "testuser", "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "UserID not found in context"})
			return
		}
		username, exists := c.Get("username")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Username not found in context"})
			return
		}
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Role not found in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"userID":   userID,
			"username": username,
			"role":     role,
		})
	})

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "No token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid token",
			token:          validToken,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
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

	// Hash passwords for test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("adminpass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash admin password: %v", err)
	}

	// Create admin user
	_, err = database.GetDB().Exec(
		"INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		1, "admin", string(hashedPassword), "admin",
	)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	hashedPassword, err = bcrypt.GenerateFromPassword([]byte("userpass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash user password: %v", err)
	}

	// Create regular user
	_, err = database.GetDB().Exec(
		"INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		2, "user", string(hashedPassword), "user",
	)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	// Generate valid JWT tokens
	adminToken, err := GenerateToken(1, "admin", "admin")
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}

	userToken, err := GenerateToken(2, "user", "user")
	if err != nil {
		t.Fatalf("Failed to generate user token: %v", err)
	}

	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(AuthMiddleware())
	router.Use(AdminMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Admin user",
			token:          adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Regular user",
			token:          userToken,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
