package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
	"github.com/stretchr/testify/assert"
)

func setupUserTestDB(t *testing.T) {
	err := database.InitTestDB()
	assert.NoError(t, err)

	// Insert test data
	_, err = database.GetDB().Exec("INSERT INTO users (id, username, password, role) VALUES (1, 'testuser', 'password', 'user')")
	assert.NoError(t, err)
	_, err = database.GetDB().Exec("INSERT INTO users (id, username, password, role) VALUES (2, 'root', 'password', 'root')")
	assert.NoError(t, err)
	_, err = database.GetDB().Exec("INSERT INTO users (id, username, password, role) VALUES (3, 'admin', 'password', 'user')")
	assert.NoError(t, err)
}

func TestGetUsers(t *testing.T) {
	setupUserTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		role           string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Root user can get all users",
			userID:         2,
			role:           "root",
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "Regular user cannot get users",
			userID:         1,
			role:           "user",
			expectedStatus: http.StatusForbidden,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)
			c.Request = httptest.NewRequest("GET", "/api/users", nil)

			GetUsers(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string][]models.User
				json.Unmarshal(w.Body.Bytes(), &response)
				users := response["users"]
				assert.Equal(t, tt.expectedCount, len(users))
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	setupUserTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		role           string
		targetUserID   string
		expectedStatus int
	}{
		{
			name:           "Root user can get any user",
			userID:         2,
			role:           "root",
			targetUserID:   "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Regular user cannot get users",
			userID:         1,
			role:           "user",
			targetUserID:   "1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid user ID",
			userID:         2,
			role:           "root",
			targetUserID:   "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "User not found",
			userID:         2,
			role:           "root",
			targetUserID:   "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)
			c.Params = gin.Params{{Key: "id", Value: tt.targetUserID}}
			c.Request = httptest.NewRequest("GET", "/api/users/"+tt.targetUserID, nil)

			GetUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateUser(t *testing.T) {
	setupUserTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		role           string
		newUser        models.User
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "Root user can create user",
			userID: 2,
			role:   "root",
			newUser: models.User{
				Username: "newuser",
				Password: "password",
				Role:     "user",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Root user can create root user",
			userID: 2,
			role:   "root",
			newUser: models.User{
				Username: "newroot",
				Password: "password",
				Role:     "root",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Invalid role",
			userID: 2,
			role:   "root",
			newUser: models.User{
				Username: "invaliduser",
				Password: "password",
				Role:     "invalidrole",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Role must be either 'user' or 'root'",
		},
		{
			name:   "Regular user cannot create users",
			userID: 1,
			role:   "user",
			newUser: models.User{
				Username: "shouldnotwork",
				Password: "password",
				Role:     "user",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "Duplicate username",
			userID: 2,
			role:   "root",
			newUser: models.User{
				Username: "testuser", // Already exists
				Password: "password",
				Role:     "user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Username already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)

			// Create request body with password field included
			requestBody := map[string]interface{}{
				"username": tt.newUser.Username,
				"password": tt.newUser.Password,
				"role":     tt.newUser.Role,
			}
			body, _ := json.Marshal(requestBody)
			c.Request = httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			CreateUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	setupUserTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		role           string
		targetUserID   string
		updateUser     models.User
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "Root user can update other users",
			userID:       2,
			role:         "root",
			targetUserID: "1",
			updateUser: models.User{
				Username: "updateduser",
				Role:     "user",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "Root user cannot update themselves",
			userID:       2,
			role:         "root",
			targetUserID: "2",
			updateUser: models.User{
				Username: "shouldnotwork",
				Role:     "user",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Cannot modify your own account",
		},
		{
			name:         "Regular user cannot update users",
			userID:       1,
			role:         "user",
			targetUserID: "3",
			updateUser: models.User{
				Username: "shouldnotwork",
				Role:     "user",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:         "Invalid user ID",
			userID:       2,
			role:         "root",
			targetUserID: "invalid",
			updateUser: models.User{
				Username: "test",
				Role:     "user",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "User not found",
			userID:       2,
			role:         "root",
			targetUserID: "999",
			updateUser: models.User{
				Username: "test",
				Role:     "user",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)
			c.Params = gin.Params{{Key: "id", Value: tt.targetUserID}}

			body, _ := json.Marshal(tt.updateUser)
			c.Request = httptest.NewRequest("PUT", "/api/users/"+tt.targetUserID, bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			UpdateUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	setupUserTestDB(t)
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userID         int64
		role           string
		targetUserID   string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Root user can delete other users",
			userID:         2,
			role:           "root",
			targetUserID:   "3",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Root user cannot delete themselves",
			userID:         2,
			role:           "root",
			targetUserID:   "2",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Cannot delete your own account",
		},
		{
			name:           "Regular user cannot delete users",
			userID:         1,
			role:           "user",
			targetUserID:   "3",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid user ID",
			userID:         2,
			role:           "root",
			targetUserID:   "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "User not found",
			userID:         2,
			role:           "root",
			targetUserID:   "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Set("userID", tt.userID)
			c.Set("role", tt.role)
			c.Params = gin.Params{{Key: "id", Value: tt.targetUserID}}
			c.Request = httptest.NewRequest("DELETE", "/api/users/"+tt.targetUserID, nil)

			DeleteUser(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}
