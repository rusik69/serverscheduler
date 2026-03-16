package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/logger"
)

// RequestIDMiddleware generates a request ID and sets it in context and response header
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := generateRequestID()
		c.Request = c.Request.WithContext(logger.WithRequestID(c.Request.Context(), id))
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// RequestLoggingMiddleware logs method, path, status, duration, request_id, client IP
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		l := logger.FromContext(c.Request.Context())
		l.Info("request",
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"client_ip", clientIP,
		)
	}
}

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
