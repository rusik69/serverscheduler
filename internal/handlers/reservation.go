package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
)

// CreateReservation handles reservation creation
func CreateReservation(c *gin.Context) {
	var reservation models.Reservation
	if err := c.ShouldBindJSON(&reservation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if server exists and is available
	var serverStatus string
	err := database.GetDB().QueryRow("SELECT status FROM servers WHERE id = ?", reservation.ServerID).Scan(&serverStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check server status"})
		return
	}

	if serverStatus != "available" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is not available"})
		return
	}

	// Check for overlapping reservations
	var count int
	err = database.GetDB().QueryRow(`
		SELECT COUNT(*) FROM reservations 
		WHERE server_id = ? AND status = 'active' AND 
		((start_time <= ? AND end_time >= ?) OR 
		(start_time <= ? AND end_time >= ?) OR 
		(start_time >= ? AND end_time <= ?))`,
		reservation.ServerID,
		reservation.StartTime, reservation.StartTime,
		reservation.EndTime, reservation.EndTime,
		reservation.StartTime, reservation.EndTime,
	).Scan(&count)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for overlapping reservations"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is already reserved for this time period"})
		return
	}

	// Create reservation
	result, err := database.GetDB().Exec(`
		INSERT INTO reservations (server_id, user_id, start_time, end_time, status) 
		VALUES (?, ?, ?, ?, ?)`,
		reservation.ServerID, userID, reservation.StartTime, reservation.EndTime, "active",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	// Update server status
	_, err = database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "reserved", reservation.ServerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server status"})
		return
	}

	id, _ := result.LastInsertId()
	reservation.ID = id
	reservation.UserID = userID.(int64)
	reservation.Status = "active"

	c.JSON(http.StatusCreated, reservation)
}

// GetReservations returns all reservations for the current user
func GetReservations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	rows, err := database.GetDB().Query(`
		SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, s.name as server_name 
		FROM reservations r 
		JOIN servers s ON r.server_id = s.id 
		WHERE r.user_id = ?`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		var serverName string
		if err := rows.Scan(
			&reservation.ID,
			&reservation.ServerID,
			&reservation.UserID,
			&reservation.StartTime,
			&reservation.EndTime,
			&reservation.Status,
			&serverName,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan reservation"})
			return
		}
		reservation.ServerName = serverName
		reservations = append(reservations, reservation)
	}

	c.JSON(http.StatusOK, reservations)
}

// GetReservation returns a specific reservation
func GetReservation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var reservation models.Reservation
	var serverName string
	err = database.GetDB().QueryRow(`
		SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, s.name as server_name 
		FROM reservations r 
		JOIN servers s ON r.server_id = s.id 
		WHERE r.id = ? AND r.user_id = ?`,
		id, userID,
	).Scan(
		&reservation.ID,
		&reservation.ServerID,
		&reservation.UserID,
		&reservation.StartTime,
		&reservation.EndTime,
		&reservation.Status,
		&serverName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservation"})
		return
	}

	reservation.ServerName = serverName
	c.JSON(http.StatusOK, reservation)
}

// CancelReservation handles reservation cancellation
func CancelReservation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get reservation details
	var serverID int64
	var status string
	err = database.GetDB().QueryRow(`
		SELECT server_id, status FROM reservations 
		WHERE id = ? AND user_id = ?`,
		id, userID,
	).Scan(&serverID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservation"})
		return
	}

	if status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation is not active"})
		return
	}

	// Update reservation status
	_, err = database.GetDB().Exec(`
		UPDATE reservations SET status = ? WHERE id = ?`,
		"cancelled", id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel reservation"})
		return
	}

	// Update server status
	_, err = database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "available", serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation cancelled successfully"})
}

// CleanupExpiredReservations cleans up expired reservations
func CleanupExpiredReservations() {
	now := time.Now()
	rows, err := database.GetDB().Query(`
		SELECT id, server_id FROM reservations 
		WHERE status = 'active' AND end_time < ?`,
		now,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, serverID int64
		if err := rows.Scan(&id, &serverID); err != nil {
			continue
		}

		// Update reservation status
		_, err = database.GetDB().Exec(`
			UPDATE reservations SET status = ? WHERE id = ?`,
			"expired", id,
		)
		if err != nil {
			continue
		}

		// Update server status
		_, err = database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "available", serverID)
		if err != nil {
			continue
		}
	}
}
