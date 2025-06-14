package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"github.com/rusik69/serverscheduler/internal/models"
)

// HandlerSet contains all the handlers needed for testing
type HandlerSet struct {
	Register          func(*gin.Context)
	Login             func(*gin.Context)
	CreateServer      func(*gin.Context)
	ListServers       func(*gin.Context)
	CreateReservation func(*gin.Context)
	ListReservations  func(*gin.Context)
	CancelReservation func(*gin.Context)
	ListUsers         func(*gin.Context)
}

// SetupTestRouter creates a new router for testing
func SetupTestRouter(handlers HandlerSet) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Public routes
	router.POST("/api/auth/register", handlers.Register)
	router.POST("/api/auth/login", handlers.Login)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("/servers", handlers.CreateServer)
		protected.GET("/servers", handlers.ListServers)
		protected.POST("/reservations", handlers.CreateReservation)
		protected.GET("/reservations", handlers.ListReservations)
		protected.DELETE("/reservations/:id", handlers.CancelReservation)
		protected.GET("/users", handlers.ListUsers)
	}

	return router
}

// GetAuthToken retrieves an authentication token for a user
func GetAuthToken(t *testing.T, router *gin.Engine, username, password string) string {
	loginPayload := models.LoginRequest{
		Username: username,
		Password: password,
	}
	jsonData, _ := json.Marshal(loginPayload)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Failed to get auth token: %v", w.Code)
		return ""
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
		return ""
	}

	return response["token"]
}
