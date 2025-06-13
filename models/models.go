package models

import (
	"time"
)

// User represents a system user
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"` // Password is not exposed in JSON
	Role     string `json:"role"`
}

// Server represents a physical or virtual server
type Server struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // available, reserved, maintenance
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Reservation represents a server reservation
type Reservation struct {
	ID        int64     `json:"id"`
	ServerID  int64     `json:"server_id"`
	UserID    int64     `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"` // active, cancelled, completed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ReservationRequest represents the reservation request payload
type ReservationRequest struct {
	ServerID  int64     `json:"server_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

// ServerRequest represents the server creation request payload
type ServerRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
