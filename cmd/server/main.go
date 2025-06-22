package main

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/handlers"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultRootUsername = "root"
	passwordLength      = 16
)

// setupLogger configures structured logging
func setupLogger() *slog.Logger {
	// Create a text handler with timestamp
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format timestamp
			if a.Key == slog.TimeKey {
				a.Value = slog.StringValue(time.Now().Format("2006-01-02 15:04:05"))
			}
			return a
		},
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

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

// resetRootPassword resets the root user password
func resetRootPassword() error {
	// Use environment password or generate random password
	var password string
	if envPassword := os.Getenv("ROOT_PASSWORD"); envPassword != "" {
		password = envPassword
		slog.Info("Using ROOT_PASSWORD from environment variable for reset")
	} else {
		var err error
		password, err = generateRandomPassword(passwordLength)
		if err != nil {
			return fmt.Errorf("failed to generate random password: %v", err)
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update root user password
	_, err = database.GetDB().Exec(
		"UPDATE users SET password = ? WHERE username = ?",
		string(hashedPassword), defaultRootUsername,
	)
	if err != nil {
		return fmt.Errorf("failed to update root password: %v", err)
	}

	slog.Info("Root password reset successfully", "username", defaultRootUsername, "password", password)
	return nil
}

func createRootUser() error {
	// Check if root user exists
	var exists bool
	err := database.GetDB().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", defaultRootUsername).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check root user: %v", err)
	}

	if exists {
		slog.Info("Root user already exists", "username", defaultRootUsername)

		// Check if password reset is requested
		if os.Getenv("RESET_ROOT_PASSWORD") == "true" {
			slog.Info("RESET_ROOT_PASSWORD=true detected, resetting root password")
			return resetRootPassword()
		}

		// Check if password is stored in environment variable
		if envPassword := os.Getenv("ROOT_PASSWORD"); envPassword != "" {
			slog.Info("Root password is available in environment variable", "password", envPassword)
		} else {
			slog.Info("Root password is not available (stored as hash in database)")
			slog.Info("To reset root password, set RESET_ROOT_PASSWORD=true environment variable")
			slog.Info("To use a specific password, set ROOT_PASSWORD environment variable")
		}
		return nil
	}

	// Use environment password or generate random password
	var password string
	if envPassword := os.Getenv("ROOT_PASSWORD"); envPassword != "" {
		password = envPassword
		slog.Info("Using ROOT_PASSWORD from environment variable")
	} else {
		var err error
		password, err = generateRandomPassword(passwordLength)
		if err != nil {
			return fmt.Errorf("failed to generate random password: %v", err)
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Create root user
	_, err = database.GetDB().Exec(
		"INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		defaultRootUsername, string(hashedPassword), "root",
	)
	if err != nil {
		return fmt.Errorf("failed to create root user: %v", err)
	}

	slog.Info("Root user created successfully", "username", defaultRootUsername, "password", password)
	return nil
}

func main() {
	// Set up structured logging
	logger := setupLogger()
	logger.Info("Starting ServerScheduler")

	// Initialize database
	err := database.Init()
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	logger.Info("Database initialized successfully")

	// Run database migrations
	err = database.RunMigrations()
	if err != nil {
		logger.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Database migrations completed successfully")

	// Create root user if it doesn't exist
	err = createRootUser()
	if err != nil {
		logger.Warn("Failed to check root user", "error", err)
	}

	// Set up router
	r := gin.Default()

	// Add custom logging middleware
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			return fmt.Sprintf("[%s] %s %s %d %s %s\n",
				params.TimeStamp.Format("2006-01-02 15:04:05"),
				params.Method,
				params.Path,
				params.StatusCode,
				params.Latency,
				params.ClientIP,
			)
		},
	}))

	// Simple CORS middleware - must be first
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Max-Age", "43200")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})

	// Serve static files from the frontend/dist directory
	r.Static("/static", "./frontend/dist/static")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// API routes
	api := r.Group("/api")
	{
		// Health check endpoint
		api.GET("/health", func(c *gin.Context) {
			// Add CORS headers directly for testing
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/user", handlers.GetUserInfo)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Server routes
			servers := protected.Group("/servers")
			{
				servers.GET("", handlers.GetServers)
				servers.GET("/:id", handlers.GetServer)
				servers.POST("", handlers.CreateServer)
				servers.PUT("/:id", handlers.UpdateServer)
				servers.DELETE("/:id", handlers.DeleteServer)
			}

			// Reservation routes
			reservations := protected.Group("/reservations")
			{
				reservations.GET("", handlers.GetReservations)
				reservations.GET("/:id", handlers.GetReservation)
				reservations.POST("", handlers.CreateReservation)
				reservations.PUT("/:id", handlers.UpdateReservation)
				reservations.DELETE("/:id", handlers.CancelReservation)
			}

			// User management routes (root only)
			users := protected.Group("/users")
			users.Use(middleware.RootMiddleware())
			{
				users.GET("", handlers.GetUsers)
				users.GET("/:id", handlers.GetUser)
				users.POST("", handlers.CreateUser)
				users.PUT("/:id", handlers.UpdateUser)
				users.DELETE("/:id", handlers.DeleteUser)
			}
		}
	}

	// Handle all other routes by serving the index.html
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Info("Server starting", "port", port, "url", fmt.Sprintf("http://localhost:%s", port))
	if err := r.Run(":" + port); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
