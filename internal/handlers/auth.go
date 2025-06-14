package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert user into database
	_, err = database.GetDB().Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		user.Username, string(hashedPassword), "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// Login handles user login
func Login(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from database
	var user models.User
	var hashedPassword string
	err := database.GetDB().QueryRow("SELECT id, username, password, role FROM users WHERE username = ?",
		loginRequest.Username).Scan(&user.ID, &user.Username, &hashedPassword, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(loginRequest.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key")) // In production, use environment variable
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

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
	userID, exists := c.Get("user_id")
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
