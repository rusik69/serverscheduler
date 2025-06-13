package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rusik69/serverscheduler/database"
	"github.com/rusik69/serverscheduler/models"
	"github.com/rusik69/serverscheduler/testutils"
)

func TestRegister(t *testing.T) {
	// Initialize test database
	if err := database.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := testutils.SetupTestRouter(testutils.HandlerSet{
		Register: Register,
		Login:    Login,
	})

	tests := []struct {
		name       string
		payload    models.RegisterRequest
		wantStatus int
	}{
		{
			name: "valid registration",
			payload: models.RegisterRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "duplicate username",
			payload: models.RegisterRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "empty username",
			payload: models.RegisterRequest{
				Username: "",
				Password: "testpass123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty password",
			payload: models.RegisterRequest{
				Username: "testuser2",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusCreated {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["token"]; !exists {
					t.Error("Register() response missing token")
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	// Initialize test database
	if err := database.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := testutils.SetupTestRouter(testutils.HandlerSet{
		Register: Register,
		Login:    Login,
	})

	// Register a test user first
	registerPayload := models.RegisterRequest{
		Username: "testuser",
		Password: "testpass123",
	}
	jsonData, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	tests := []struct {
		name       string
		payload    models.LoginRequest
		wantStatus int
	}{
		{
			name: "valid login",
			payload: models.LoginRequest{
				Username: "testuser",
				Password: "testpass123",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid password",
			payload: models.LoginRequest{
				Username: "testuser",
				Password: "wrongpass",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "non-existent user",
			payload: models.LoginRequest{
				Username: "nonexistent",
				Password: "testpass123",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "empty username",
			payload: models.LoginRequest{
				Username: "",
				Password: "testpass123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty password",
			payload: models.LoginRequest{
				Username: "testuser",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["token"]; !exists {
					t.Error("Login() response missing token")
				}
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	// Initialize test database
	if err := database.InitTestDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	defer database.CleanupTestDB()

	router := testutils.SetupTestRouter(testutils.HandlerSet{
		Register:  Register,
		Login:     Login,
		ListUsers: ListUsers,
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
		wantStatus int
	}{
		{
			name:       "list users as root",
			token:      rootToken,
			wantStatus: http.StatusOK,
		},
		{
			name:       "list users as regular user",
			token:      regularToken,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "list users without auth",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/users", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ListUsers() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string][]models.User
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if len(response["users"]) == 0 {
					t.Error("ListUsers() returned empty list")
				}
				// Verify that passwords are not included in the response
				for _, user := range response["users"] {
					if user.Password != "" {
						t.Error("ListUsers() included password in response")
					}
				}
			}
		})
	}
}
