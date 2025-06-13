package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/database"
	"github.com/rusik69/serverscheduler/handlers"
	"github.com/rusik69/serverscheduler/middleware"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultRootUsername = "root"
	passwordLength      = 16
)

// generateRandomPassword creates a secure random password
func generateRandomPassword(length int) (string, error) {
	const (
		lowerChars = "abcdefghijklmnopqrstuvwxyz"
		upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numChars   = "0123456789"
		specChars  = "!@#$%^&*"
		allChars   = lowerChars + upperChars + numChars + specChars
	)

	// Ensure at least one character from each category
	password := make([]byte, length)

	// Add one character from each category
	password[0] = lowerChars[getRandomInt(len(lowerChars))]
	password[1] = upperChars[getRandomInt(len(upperChars))]
	password[2] = numChars[getRandomInt(len(numChars))]
	password[3] = specChars[getRandomInt(len(specChars))]

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password[i] = allChars[getRandomInt(len(allChars))]
	}

	// Shuffle the password
	for i := len(password) - 1; i > 0; i-- {
		j := getRandomInt(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}

// getRandomInt returns a random integer in the range [0, max)
func getRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Fallback to a less secure but still random number
		return int(n.Int64())
	}
	return int(n.Int64())
}

func createRootUser() error {
	// Check if root user exists
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", defaultRootUsername).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check root user: %v", err)
	}

	if exists {
		log.Println("Root user already exists")
		return nil
	}

	// Generate random password
	password, err := generateRandomPassword(passwordLength)
	if err != nil {
		return fmt.Errorf("failed to generate random password: %v", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Create root user
	_, err = database.DB.Exec(
		"INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		defaultRootUsername, string(hashedPassword), "root",
	)
	if err != nil {
		return fmt.Errorf("failed to create root user: %v", err)
	}

	log.Printf("Root user created successfully!\nUsername: %s\nPassword: %s\n", defaultRootUsername, password)
	return nil
}

func main() {
	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Create root user if it doesn't exist
	if err := createRootUser(); err != nil {
		log.Printf("Warning: %v\n", err)
	}

	// Create Gin router
	r := gin.Default()

	// Public routes
	r.POST("/api/auth/register", handlers.Register)
	r.POST("/api/auth/login", handlers.Login)

	// Protected routes
	authorized := r.Group("/api")
	authorized.Use(middleware.AuthMiddleware())
	{
		// Server routes
		authorized.GET("/servers", handlers.ListServers)
		authorized.POST("/servers", handlers.CreateServer)

		// Reservation routes
		authorized.GET("/reservations", handlers.ListReservations)
		authorized.POST("/reservations", handlers.CreateReservation)
		authorized.DELETE("/reservations/:id", handlers.CancelReservation)
	}

	// Start server
	log.Println("Server starting on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
