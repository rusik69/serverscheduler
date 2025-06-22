package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// GetUsers returns all users (root only)
func GetUsers(c *gin.Context) {
	username, _ := c.Get("username")
	slog.Info("Fetching all users", "admin_user", username, "client_ip", c.ClientIP())

	rows, err := database.GetDB().Query("SELECT id, username, role FROM users ORDER BY id")
	if err != nil {
		slog.Error("Failed to query users", "admin_user", username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			slog.Error("Failed to scan user row", "admin_user", username, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user"})
			return
		}
		users = append(users, user)
	}

	slog.Info("Users fetched successfully", "admin_user", username, "count", len(users), "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUser returns a specific user (root only)
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Warn("Invalid user ID in get request", "id", idStr, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	adminUser, _ := c.Get("username")
	slog.Info("Fetching user", "user_id", id, "admin_user", adminUser, "client_ip", c.ClientIP())

	var user models.User
	err = database.GetDB().QueryRow("SELECT id, username, role FROM users WHERE id = ?", id).Scan(
		&user.ID, &user.Username, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("User not found", "user_id", id, "admin_user", adminUser)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		slog.Error("Failed to query user", "user_id", id, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	slog.Info("User fetched successfully", "user_id", id, "username", user.Username, "admin_user", adminUser, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user (root only)
func CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid user creation request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminUser, _ := c.Get("username")
	slog.Info("User creation attempt", "new_username", req.Username, "role", req.Role, "admin_user", adminUser, "client_ip", c.ClientIP())

	// Validate role
	if req.Role != "user" && req.Role != "root" {
		slog.Warn("User creation failed - invalid role", "role", req.Role, "admin_user", adminUser)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'user' or 'root'"})
		return
	}

	// Check if username already exists
	var exists bool
	err := database.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", req.Username).Scan(&exists)
	if err != nil {
		slog.Error("Failed to check username existence", "username", req.Username, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username"})
		return
	}

	if exists {
		slog.Warn("User creation failed - username already exists", "username", req.Username, "admin_user", adminUser)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password", "username", req.Username, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	result, err := database.GetDB().Exec(
		"INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		req.Username, string(hashedPassword), req.Role,
	)
	if err != nil {
		slog.Error("Failed to create user in database", "username", req.Username, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	id, _ := result.LastInsertId()
	user := models.User{
		ID:       id,
		Username: req.Username,
		Role:     req.Role,
	}

	slog.Info("User created successfully", "user_id", id, "username", req.Username, "role", req.Role, "admin_user", adminUser, "client_ip", c.ClientIP())
	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates an existing user (root only)
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Warn("Invalid user ID in update request", "id", idStr, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid user update request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminUser, _ := c.Get("username")
	adminUserID, _ := c.Get("userID")
	slog.Info("User update attempt", "user_id", id, "admin_user", adminUser, "client_ip", c.ClientIP())

	// Check if user exists
	var currentUser models.User
	err = database.GetDB().QueryRow("SELECT id, username, role FROM users WHERE id = ?", id).Scan(
		&currentUser.ID, &currentUser.Username, &currentUser.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("User update failed - user not found", "user_id", id, "admin_user", adminUser)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		slog.Error("Failed to query user for update", "user_id", id, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user"})
		return
	}

	// Prevent admins from modifying themselves (security measure)
	if adminUserID == id {
		slog.Warn("User update failed - cannot modify own account", "user_id", id, "admin_user", adminUser)
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify your own account"})
		return
	}

	// Build update query dynamically
	var setParts []string
	var args []interface{}

	if req.Username != "" && req.Username != currentUser.Username {
		// Check if new username already exists
		var exists bool
		err := database.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? AND id != ?)", req.Username, id).Scan(&exists)
		if err != nil {
			slog.Error("Failed to check username existence", "username", req.Username, "admin_user", adminUser, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check username"})
			return
		}
		if exists {
			slog.Warn("User update failed - username already exists", "username", req.Username, "admin_user", adminUser)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
			return
		}
		setParts = append(setParts, "username = ?")
		args = append(args, req.Username)
		currentUser.Username = req.Username
	}

	if req.Password != "" {
		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("Failed to hash password", "user_id", id, "admin_user", adminUser, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		setParts = append(setParts, "password = ?")
		args = append(args, string(hashedPassword))
	}

	if req.Role != "" && req.Role != currentUser.Role {
		// Validate role
		if req.Role != "user" && req.Role != "root" {
			slog.Warn("User update failed - invalid role", "role", req.Role, "admin_user", adminUser)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'user' or 'root'"})
			return
		}
		setParts = append(setParts, "role = ?")
		args = append(args, req.Role)
		currentUser.Role = req.Role
	}

	if len(setParts) == 0 {
		slog.Info("User update - no changes provided", "user_id", id, "admin_user", adminUser)
		c.JSON(http.StatusOK, currentUser)
		return
	}

	// Update user
	query := "UPDATE users SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err = database.GetDB().Exec(query, args...)
	if err != nil {
		slog.Error("Failed to update user in database", "user_id", id, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	slog.Info("User updated successfully", "user_id", id, "username", currentUser.Username, "role", currentUser.Role, "admin_user", adminUser, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, currentUser)
}

// DeleteUser deletes a user (root only)
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Warn("Invalid user ID in delete request", "id", idStr, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	adminUser, _ := c.Get("username")
	adminUserID, _ := c.Get("userID")
	slog.Info("User deletion attempt", "user_id", id, "admin_user", adminUser, "client_ip", c.ClientIP())

	// Check if user exists
	var user models.User
	err = database.GetDB().QueryRow("SELECT id, username, role FROM users WHERE id = ?", id).Scan(
		&user.ID, &user.Username, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("User deletion failed - user not found", "user_id", id, "admin_user", adminUser)
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		slog.Error("Failed to query user for deletion", "user_id", id, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query user"})
		return
	}

	// Prevent admins from deleting themselves
	if adminUserID == id {
		slog.Warn("User deletion failed - cannot delete own account", "user_id", id, "admin_user", adminUser)
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Check for active reservations
	var reservationCount int
	err = database.GetDB().QueryRow("SELECT COUNT(*) FROM reservations WHERE user_id = ? AND status = 'active'", id).Scan(&reservationCount)
	if err != nil {
		slog.Error("Failed to check user reservations", "user_id", id, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user reservations"})
		return
	}

	if reservationCount > 0 {
		slog.Warn("User deletion failed - user has active reservations", "user_id", id, "username", user.Username, "reservation_count", reservationCount, "admin_user", adminUser)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete user with active reservations"})
		return
	}

	// Delete user
	_, err = database.GetDB().Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		slog.Error("Failed to delete user from database", "user_id", id, "admin_user", adminUser, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	slog.Info("User deleted successfully", "user_id", id, "username", user.Username, "admin_user", adminUser, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
