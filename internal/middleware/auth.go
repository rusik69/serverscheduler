package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your-secret-key") // In production, use environment variable

// Claims represents the JWT claims
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token
func GenerateToken(userID int64, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// AuthMiddleware is a middleware for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			slog.Warn("Authentication failed - missing authorization header", "path", c.Request.URL.Path, "client_ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			slog.Warn("Authentication failed - invalid header format", "path", c.Request.URL.Path, "client_ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			slog.Warn("Authentication failed - invalid token", "path", c.Request.URL.Path, "client_ip", c.ClientIP(), "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		slog.Info("Authentication successful", "user_id", claims.UserID, "username", claims.Username, "role", claims.Role, "path", c.Request.URL.Path, "client_ip", c.ClientIP())
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminMiddleware is a middleware to check if the user is an admin
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		username, _ := c.Get("username")
		if !exists || role != "admin" {
			slog.Warn("Admin access denied", "username", username, "role", role, "path", c.Request.URL.Path, "client_ip", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		slog.Info("Admin access granted", "username", username, "path", c.Request.URL.Path, "client_ip", c.ClientIP())
		c.Next()
	}
}

// RootMiddleware is a middleware to check if the user is root
func RootMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		username, _ := c.Get("username")
		if !exists || role != "root" {
			slog.Warn("Root access denied", "username", username, "role", role, "path", c.Request.URL.Path, "client_ip", c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Root access required"})
			c.Abort()
			return
		}
		slog.Info("Root access granted", "username", username, "path", c.Request.URL.Path, "client_ip", c.ClientIP())
		c.Next()
	}
}
