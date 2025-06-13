package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/database"
	"github.com/rusik69/serverscheduler/middleware"
	"github.com/rusik69/serverscheduler/models"
	"github.com/rusik69/serverscheduler/testutils"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Public routes
	r.POST("/api/auth/register", Register)
	r.POST("/api/auth/login", Login)

	// Protected routes
	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.GET("/servers", ListServers)
		authorized.POST("/servers", CreateServer)
	}

	return r
}

func getAuthToken(t *testing.T, router *gin.Engine, username, password string) string {
	loginPayload := models.LoginRequest{
		Username: username,
		Password: password,
	}
	jsonData, _ := json.Marshal(loginPayload)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["token"]
}

func TestCreateServer(t *testing.T) {
	// Initialize test database
	if err := database.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := testutils.SetupTestRouter(testutils.HandlerSet{
		Register:     Register,
		Login:        Login,
		CreateServer: CreateServer,
	})

	// Register root user
	registerPayload := models.RegisterRequest{
		Username: "root",
		Password: "rootpass123",
	}
	jsonData, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get root token
	rootToken := testutils.GetAuthToken(t, router, "root", "rootpass123")

	// Register regular user
	registerPayload = models.RegisterRequest{
		Username: "regular",
		Password: "regularpass123",
	}
	jsonData, _ = json.Marshal(registerPayload)
	req, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get regular user token
	regularToken := testutils.GetAuthToken(t, router, "regular", "regularpass123")

	tests := []struct {
		name       string
		token      string
		payload    models.Server
		wantStatus int
	}{
		{
			name:  "create server as root",
			token: rootToken,
			payload: models.Server{
				Name:   "test-server-1",
				Status: "available",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:  "create server as regular user",
			token: regularToken,
			payload: models.Server{
				Name:   "test-server-2",
				Status: "available",
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:  "create server without auth",
			token: "",
			payload: models.Server{
				Name:   "test-server-3",
				Status: "available",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "create server with empty name",
			token: rootToken,
			payload: models.Server{
				Name:   "",
				Status: "available",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateServer() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestListServers(t *testing.T) {
	// Initialize test database
	if err := database.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := testutils.SetupTestRouter(testutils.HandlerSet{
		Register:     Register,
		Login:        Login,
		CreateServer: CreateServer,
		ListServers:  ListServers,
	})

	// Register root user
	registerPayload := models.RegisterRequest{
		Username: "root",
		Password: "rootpass123",
	}
	jsonData, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get root token
	rootToken := testutils.GetAuthToken(t, router, "root", "rootpass123")

	// Create a test server
	serverPayload := models.Server{
		Name:   "test-server",
		Status: "available",
	}
	jsonData, _ = json.Marshal(serverPayload)
	req, _ = http.NewRequest("POST", "/api/servers", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+rootToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "list servers as root",
			token:      rootToken,
			wantStatus: http.StatusOK,
		},
		{
			name:       "list servers without auth",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/servers", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ListServers() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string][]models.Server
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if len(response["servers"]) == 0 {
					t.Error("ListServers() returned empty list")
				}
			}
		})
	}
}
