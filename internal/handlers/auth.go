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
