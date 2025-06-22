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

func setupServerTestDB(t *testing.T) {
	err := database.InitTestDB()
	assert.NoError(t, err)

	// Insert test data
	_, err = database.GetDB().Exec("INSERT INTO servers (id, name, status, ip_address, username, password) VALUES (1, 'test-server', 'available', '192.168.1.100', 'admin', 'password')")
	assert.NoError(t, err)
	_, err = database.GetDB().Exec("INSERT INTO servers (id, name, status) VALUES (2, 'server-2', 'reserved')")
	assert.NoError(t, err)
}

func TestCreateServer(t *testing.T) {
	setupServerTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		server         models.Server
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid server creation",
			server: models.Server{
				Name:      "new-server",
				IPAddress: "192.168.1.200",
				Username:  "admin",
				Password:  "secret",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Server with empty IP address",
			server: models.Server{
				Name:     "server-no-ip",
				Username: "admin",
				Password: "secret",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Server with IPv6 address",
			server: models.Server{
				Name:      "server-ipv6",
				IPAddress: "2001:db8::1",
				Username:  "admin",
				Password:  "secret",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid IP address",
			server: models.Server{
				Name:      "invalid-ip-server",
				IPAddress: "256.256.256.256",
				Username:  "admin",
				Password:  "secret",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid IP address format",
		},
		{
			name: "Invalid IP address format",
			server: models.Server{
				Name:      "bad-ip-server",
				IPAddress: "not-an-ip",
				Username:  "admin",
				Password:  "secret",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid IP address format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.server)
			c.Request = httptest.NewRequest("POST", "/api/servers", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			CreateServer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestGetServers(t *testing.T) {
	setupServerTestDB(t)
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/servers", nil)

	GetServers(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	servers := response["servers"].([]interface{})
	assert.Equal(t, 2, len(servers))
}

func TestGetServer(t *testing.T) {
	setupServerTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serverID       string
		expectedStatus int
	}{
		{
			name:           "Valid server ID",
			serverID:       "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid server ID",
			serverID:       "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Server not found",
			serverID:       "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.serverID}}
			c.Request = httptest.NewRequest("GET", "/api/servers/"+tt.serverID, nil)

			GetServer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUpdateServer(t *testing.T) {
	setupServerTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serverID       string
		server         models.Server
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "Valid server update",
			serverID: "1",
			server: models.Server{
				Name:      "updated-server",
				Status:    "maintenance",
				IPAddress: "192.168.1.150",
				Username:  "newadmin",
				Password:  "newpassword",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Update with invalid IP",
			serverID: "1",
			server: models.Server{
				Name:      "server-bad-ip",
				Status:    "available",
				IPAddress: "invalid-ip",
				Username:  "admin",
				Password:  "password",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid IP address format",
		},
		{
			name:           "Invalid server ID",
			serverID:       "invalid",
			server:         models.Server{Name: "test"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.serverID}}

			body, _ := json.Marshal(tt.server)
			c.Request = httptest.NewRequest("PUT", "/api/servers/"+tt.serverID, bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			UpdateServer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestDeleteServer(t *testing.T) {
	setupServerTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		serverID       string
		expectedStatus int
	}{
		{
			name:           "Valid server deletion",
			serverID:       "2",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid server ID",
			serverID:       "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.serverID}}
			c.Request = httptest.NewRequest("DELETE", "/api/servers/"+tt.serverID, nil)

			DeleteServer(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// Valid IPv4 addresses
		{"Valid IPv4 - 192.168.1.1", "192.168.1.1", true},
		{"Valid IPv4 - 10.0.0.1", "10.0.0.1", true},
		{"Valid IPv4 - 172.16.0.1", "172.16.0.1", true},
		{"Valid IPv4 - 127.0.0.1", "127.0.0.1", true},
		{"Valid IPv4 - 255.255.255.255", "255.255.255.255", true},
		{"Valid IPv4 - 0.0.0.0", "0.0.0.0", true},

		// Valid IPv6 addresses
		{"Valid IPv6 - full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"Valid IPv6 - compressed", "2001:db8:85a3::8a2e:370:7334", true},
		{"Valid IPv6 - loopback", "::1", true},
		{"Valid IPv6 - all zeros", "::", true},

		// Empty string (should be allowed)
		{"Empty string", "", true},
		{"Whitespace only", "   ", true},

		// Invalid addresses
		{"Invalid IPv4 - too high", "256.1.1.1", false},
		{"Invalid IPv4 - negative", "-1.1.1.1", false},
		{"Invalid IPv4 - incomplete", "192.168.1", false},
		{"Invalid IPv4 - too many octets", "192.168.1.1.1", false},
		{"Invalid IPv4 - letters", "192.168.a.1", false},
		{"Invalid IPv6 - too many groups", "2001:0db8:85a3:0000:0000:8a2e:0370:7334:extra", false},
		{"Invalid - random string", "not-an-ip", false},
		{"Invalid - hostname", "example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateIPAddress(tt.ip)
			if result != tt.expected {
				t.Errorf("validateIPAddress(%q) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}
