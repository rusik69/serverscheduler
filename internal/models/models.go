package models

import "time"

// User represents a user in the system
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	SSHPublicKey string   `json:"ssh_public_key,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserPublic is a user without sensitive fields
type UserPublic struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Server represents a target server for scheduling
type Server struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Hostname       string    `json:"hostname"`
	Port           int       `json:"port"`
	SSHUser        string    `json:"ssh_user"`
	SSHPrivateKey  string    `json:"-"` // never expose to API
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}

// Reservation represents a scheduled access window
type Reservation struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ServerID  int64     `json:"server_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"` // pending, active, expired, cancelled
	CreatedAt time.Time `json:"created_at"`
}

// ReservationWithDetails includes server and user info
type ReservationWithDetails struct {
	Reservation
	ServerName string `json:"server_name"`
	Username   string `json:"username"`
}
