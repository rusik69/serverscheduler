package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"github.com/rusik69/serverscheduler/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid registration request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if req.Username == "" {
		slog.Warn("Registration attempt with empty username", "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}
	if req.Password == "" {
		slog.Warn("Registration attempt with empty password", "username", req.Username, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	slog.Info("User registration attempt", "username", req.Username, "client_ip", c.ClientIP())

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash password during registration", "username", req.Username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert user into database
	_, err = database.GetDB().Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		req.Username, string(hashedPassword), "user")
	if err != nil {
		slog.Error("Failed to create user in database", "username", req.Username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	slog.Info("User registered successfully", "username", req.Username, "client_ip", c.ClientIP())
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Login handles user login
func Login(c *gin.Context) {
	var loginRequest models.LoginRequest

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		slog.Warn("Invalid login request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slog.Info("Login attempt", "username", loginRequest.Username, "client_ip", c.ClientIP())

	// Get user from database
	var user models.User
	var hashedPassword string
	err := database.GetDB().QueryRow("SELECT id, username, password, role FROM users WHERE username = ?",
		loginRequest.Username).Scan(&user.ID, &user.Username, &hashedPassword, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Login failed - user not found", "username", loginRequest.Username, "client_ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		slog.Error("Database error during login", "username", loginRequest.Username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(loginRequest.Password))
	if err != nil {
		slog.Warn("Login failed - invalid password", "username", loginRequest.Username, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token using the middleware's GenerateToken function
	tokenString, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		slog.Error("Failed to generate token", "username", loginRequest.Username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	slog.Info("Login successful", "username", loginRequest.Username, "user_id", user.ID, "role", user.Role, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// GetUserInfo returns information about the current user
func GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	err := database.GetDB().QueryRow("SELECT id, username, role FROM users WHERE id = ?",
		userID).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

// ListUsers handles listing all users (root only)
func ListUsers(c *gin.Context) {
	// Check if user is root
	role, exists := c.Get("role")
	if !exists || role != "root" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only root user can list users"})
		return
	}

	rows, err := database.GetDB().Query("SELECT id, username, role FROM users")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user"})
			return
		}
		// Don't include password in the response
		user.Password = ""
		users = append(users, user)
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// ChangePassword handles password changes for authenticated users
func ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid change password request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user info from context
	userID, exists := c.Get("userID")
	if !exists {
		slog.Warn("Password change attempt without authentication", "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		slog.Warn("Password change attempt without username in context", "user_id", userID, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		slog.Warn("Password change attempt without role in context", "user_id", userID, "username", username, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	slog.Info("Password change attempt", "user_id", userID, "username", username, "role", role, "client_ip", c.ClientIP())

	// Validate required fields
	if req.CurrentPassword == "" {
		slog.Warn("Password change failed - current password required", "user_id", userID, "username", username, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is required"})
		return
	}
	if req.NewPassword == "" {
		slog.Warn("Password change failed - new password required", "user_id", userID, "username", username, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password is required"})
		return
	}
	if len(req.NewPassword) < 6 {
		slog.Warn("Password change failed - new password too short", "user_id", userID, "username", username, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 6 characters long"})
		return
	}
	if req.NewPassword == req.CurrentPassword {
		slog.Warn("Password change failed - new password same as current", "user_id", userID, "username", username, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different from current password"})
		return
	}

	// Get current password hash from database
	var currentHashedPassword string
	err := database.GetDB().QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&currentHashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("Password change failed - user not found", "user_id", userID, "username", username, "client_ip", c.ClientIP())
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		slog.Error("Database error during password change", "user_id", userID, "username", username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(currentHashedPassword), []byte(req.CurrentPassword))
	if err != nil {
		slog.Warn("Password change failed - current password incorrect", "user_id", userID, "username", username, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Failed to hash new password", "user_id", userID, "username", username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password in database
	_, err = database.GetDB().Exec("UPDATE users SET password = ? WHERE id = ?", string(newHashedPassword), userID)
	if err != nil {
		slog.Error("Failed to update password in database", "user_id", userID, "username", username, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	slog.Info("Password changed successfully", "user_id", userID, "username", username, "role", role, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
