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
	IPAddress   string    `json:"ip_address"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Reservation represents a server reservation
type Reservation struct {
	ID             int64     `json:"id"`
	ServerID       int64     `json:"server_id"`
	ServerName     string    `json:"server_name"`
	ServerUsername string    `json:"server_username,omitempty"` // Only shown to root users
	ServerPassword string    `json:"server_password,omitempty"` // Only shown to root users
	ServerIP       string    `json:"server_ip,omitempty"`       // Server IP address
	UserID         int64     `json:"user_id"`
	Username       string    `json:"username"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         string    `json:"status"` // active, cancelled, completed
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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

// ChangePasswordRequest represents the password change request payload
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
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
