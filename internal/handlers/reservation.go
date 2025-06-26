package handlers

import (
	"database/sql"
	"log/slog"
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
		slog.Warn("Invalid reservation creation request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID and username from context
	userID, exists := c.Get("userID")
	if !exists {
		slog.Warn("Reservation creation failed - user not authenticated", "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		slog.Warn("Reservation creation failed - username not in context", "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	slog.Info("Reservation creation attempt", "user_id", userID, "username", username, "server_id", reservation.ServerID, "start_time", reservation.StartTime, "end_time", reservation.EndTime, "client_ip", c.ClientIP())

	// Validate that start time is before end time
	if !reservation.StartTime.Before(reservation.EndTime) {
		slog.Warn("Reservation creation failed - start time is not before end time", "user_id", userID, "start_time", reservation.StartTime, "end_time", reservation.EndTime, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
		return
	}

	// Validate that start time is not in the past
	now := time.Now()
	if reservation.StartTime.Before(now) {
		slog.Warn("Reservation creation failed - start time is in the past", "user_id", userID, "start_time", reservation.StartTime, "current_time", now, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time cannot be in the past"})
		return
	}

	// Check if server exists and is available
	var serverStatus, serverName string
	err := database.GetDB().QueryRow("SELECT name, status FROM servers WHERE id = ?", reservation.ServerID).Scan(&serverName, &serverStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Reservation creation failed - server not found", "server_id", reservation.ServerID, "user_id", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
			return
		}
		slog.Error("Failed to check server status", "server_id", reservation.ServerID, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check server status"})
		return
	}

	if serverStatus != "available" {
		slog.Warn("Reservation creation failed - server not available", "server_id", reservation.ServerID, "status", serverStatus, "user_id", userID)
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
		slog.Error("Failed to check for overlapping reservations", "server_id", reservation.ServerID, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for overlapping reservations"})
		return
	}

	if count > 0 {
		slog.Warn("Reservation creation failed - overlapping reservation exists", "server_id", reservation.ServerID, "user_id", userID, "overlapping_count", count)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is already reserved for this time period"})
		return
	}

	// Create reservation
	result, err := database.GetDB().Exec(`
		INSERT INTO reservations (server_id, user_id, server_name, start_time, end_time, status) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		reservation.ServerID, userID, serverName, reservation.StartTime, reservation.EndTime, "active",
	)
	if err != nil {
		slog.Error("Failed to create reservation in database", "server_id", reservation.ServerID, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	// Update server status
	_, err = database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "reserved", reservation.ServerID)
	if err != nil {
		slog.Error("Failed to update server status after reservation", "server_id", reservation.ServerID, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server status"})
		return
	}

	id, _ := result.LastInsertId()
	reservation.ID = id
	reservation.UserID = userID.(int64)
	reservation.Username = username.(string)
	reservation.ServerName = serverName
	reservation.Status = "active"

	slog.Info("Reservation created successfully", "reservation_id", reservation.ID, "server_id", reservation.ServerID, "user_id", userID, "username", username, "client_ip", c.ClientIP())
	c.JSON(http.StatusCreated, reservation)
}

// GetReservations returns reservations based on user role
func GetReservations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		slog.Warn("Reservations access failed - user not authenticated", "path", c.Request.URL.Path, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		slog.Warn("Reservations access failed - user role not in context", "user_id", userID, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not authenticated"})
		return
	}

	var rows *sql.Rows
	var err error

	// Root users can see all reservations with server credentials, regular users only see their own without credentials
	if role == "root" {
		slog.Info("Fetching all reservations with server credentials (root access)", "user_id", userID, "client_ip", c.ClientIP())
		rows, err = database.GetDB().Query(`
			SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, 
			       s.name as server_name, u.username,
			       COALESCE(s.username, '') as server_username,
			       COALESCE(s.password, '') as server_password,
			       COALESCE(s.ip_address, '') as server_ip
			FROM reservations r 
			JOIN servers s ON r.server_id = s.id 
			JOIN users u ON r.user_id = u.id
			ORDER BY r.start_time DESC`)
	} else {
		slog.Info("Fetching user reservations with server credentials", "user_id", userID, "client_ip", c.ClientIP())
		rows, err = database.GetDB().Query(`
			SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, 
			       s.name as server_name, u.username,
			       COALESCE(s.username, '') as server_username,
			       COALESCE(s.password, '') as server_password,
			       COALESCE(s.ip_address, '') as server_ip
			FROM reservations r 
			JOIN servers s ON r.server_id = s.id 
			JOIN users u ON r.user_id = u.id
			WHERE r.user_id = ?
			ORDER BY r.start_time DESC`,
			userID,
		)
	}
	if err != nil {
		slog.Error("Failed to query reservations", "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		var serverName, username, serverUsername, serverPassword, serverIP string
		if err := rows.Scan(
			&reservation.ID,
			&reservation.ServerID,
			&reservation.UserID,
			&reservation.StartTime,
			&reservation.EndTime,
			&reservation.Status,
			&serverName,
			&username,
			&serverUsername,
			&serverPassword,
			&serverIP,
		); err != nil {
			slog.Error("Failed to scan reservation row", "user_id", userID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan reservation"})
			return
		}
		reservation.ServerName = serverName
		reservation.Username = username
		reservation.ServerUsername = serverUsername
		reservation.ServerPassword = serverPassword
		reservation.ServerIP = serverIP
		reservations = append(reservations, reservation)
	}

	slog.Info("Reservations fetched successfully", "user_id", userID, "count", len(reservations), "show_credentials", true, "role", role, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, reservations)
}

// GetReservation returns a specific reservation
func GetReservation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Warn("Invalid reservation ID in get request", "id", idStr, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		slog.Warn("Reservation access failed - user not authenticated", "reservation_id", id, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		slog.Warn("Reservation access failed - user role not in context", "user_id", userID, "reservation_id", id, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not authenticated"})
		return
	}

	var reservation models.Reservation
	var serverName, username, serverUsername, serverPassword, serverIP string

	// Root users can view any reservation with server credentials, regular users only their own without credentials
	if role == "root" {
		slog.Info("Fetching reservation with server credentials (root access)", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
		err = database.GetDB().QueryRow(`
			SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, 
			       s.name as server_name, u.username,
			       COALESCE(s.username, '') as server_username,
			       COALESCE(s.password, '') as server_password,
			       COALESCE(s.ip_address, '') as server_ip
			FROM reservations r 
			JOIN servers s ON r.server_id = s.id 
			JOIN users u ON r.user_id = u.id
			WHERE r.id = ?`,
			id,
		).Scan(
			&reservation.ID,
			&reservation.ServerID,
			&reservation.UserID,
			&reservation.StartTime,
			&reservation.EndTime,
			&reservation.Status,
			&serverName,
			&username,
			&serverUsername,
			&serverPassword,
			&serverIP,
		)
	} else {
		slog.Info("Fetching user reservation with server credentials", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
		err = database.GetDB().QueryRow(`
			SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, 
			       s.name as server_name, u.username,
			       COALESCE(s.username, '') as server_username,
			       COALESCE(s.password, '') as server_password,
			       COALESCE(s.ip_address, '') as server_ip
			FROM reservations r 
			JOIN servers s ON r.server_id = s.id 
			JOIN users u ON r.user_id = u.id
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
			&username,
			&serverUsername,
			&serverPassword,
			&serverIP,
		)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservation"})
		return
	}

	reservation.ServerName = serverName
	reservation.Username = username
	reservation.ServerUsername = serverUsername
	reservation.ServerPassword = serverPassword
	reservation.ServerIP = serverIP
	c.JSON(http.StatusOK, reservation)
}

// UpdateReservation handles reservation updates
func UpdateReservation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Warn("Invalid reservation ID in update request", "id", idStr, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	var updatedReservation models.Reservation
	if err := c.ShouldBindJSON(&updatedReservation); err != nil {
		slog.Warn("Invalid reservation update request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		slog.Warn("Reservation update failed - user not authenticated", "reservation_id", id, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		slog.Warn("Reservation update failed - username not in context", "reservation_id", id, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	slog.Info("Reservation update attempt", "reservation_id", id, "user_id", userID, "username", username, "new_server_id", updatedReservation.ServerID, "new_start_time", updatedReservation.StartTime, "new_end_time", updatedReservation.EndTime, "client_ip", c.ClientIP())

	// Validate that start time is before end time
	if !updatedReservation.StartTime.Before(updatedReservation.EndTime) {
		slog.Warn("Reservation update failed - start time is not before end time", "reservation_id", id, "user_id", userID, "start_time", updatedReservation.StartTime, "end_time", updatedReservation.EndTime, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
		return
	}

	// Validate that start time is not in the past
	now := time.Now()
	if updatedReservation.StartTime.Before(now) {
		slog.Warn("Reservation update failed - start time is in the past", "reservation_id", id, "user_id", userID, "start_time", updatedReservation.StartTime, "current_time", now, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time cannot be in the past"})
		return
	}

	// Get user role to determine access level
	role, exists := c.Get("role")
	if !exists {
		slog.Warn("Reservation update failed - user role not in context", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not authenticated"})
		return
	}

	// Check if reservation exists - root can access any, user only their own
	var currentReservation models.Reservation
	var query string
	var args []interface{}

	if role == "root" {
		query = `SELECT id, server_id, user_id, start_time, end_time, status FROM reservations WHERE id = ?`
		args = []interface{}{id}
		slog.Info("Reservation update check (root access)", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
	} else {
		query = `SELECT id, server_id, user_id, start_time, end_time, status FROM reservations WHERE id = ? AND user_id = ?`
		args = []interface{}{id, userID}
		slog.Info("Reservation update check (user access)", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
	}

	err = database.GetDB().QueryRow(query, args...,
	).Scan(
		&currentReservation.ID,
		&currentReservation.ServerID,
		&currentReservation.UserID,
		&currentReservation.StartTime,
		&currentReservation.EndTime,
		&currentReservation.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Reservation update failed - reservation not found or not owned by user", "reservation_id", id, "user_id", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}
		slog.Error("Failed to query reservation for update", "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query reservation"})
		return
	}

	// Check if reservation can be updated (not cancelled)
	if currentReservation.Status == "cancelled" {
		slog.Warn("Reservation update failed - reservation is cancelled", "reservation_id", id, "user_id", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot update cancelled reservation"})
		return
	}

	// Check if new server exists and is available (if server is being changed)
	var serverName string
	if updatedReservation.ServerID != currentReservation.ServerID {
		var serverStatus string
		err = database.GetDB().QueryRow("SELECT name, status FROM servers WHERE id = ?", updatedReservation.ServerID).Scan(&serverName, &serverStatus)
		if err != nil {
			if err == sql.ErrNoRows {
				slog.Warn("Reservation update failed - new server not found", "server_id", updatedReservation.ServerID, "reservation_id", id, "user_id", userID)
				c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
				return
			}
			slog.Error("Failed to check new server status", "server_id", updatedReservation.ServerID, "reservation_id", id, "user_id", userID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check server status"})
			return
		}

		if serverStatus != "available" && serverStatus != "reserved" {
			slog.Warn("Reservation update failed - new server not available", "server_id", updatedReservation.ServerID, "status", serverStatus, "reservation_id", id, "user_id", userID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server is not available"})
			return
		}
	} else {
		// Get current server name if server is not changing
		err = database.GetDB().QueryRow("SELECT name FROM servers WHERE id = ?", updatedReservation.ServerID).Scan(&serverName)
		if err != nil {
			slog.Error("Failed to get current server name", "server_id", updatedReservation.ServerID, "reservation_id", id, "user_id", userID, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get server information"})
			return
		}
	}

	// Check for overlapping reservations (excluding current reservation)
	var count int
	err = database.GetDB().QueryRow(`
		SELECT COUNT(*) FROM reservations 
		WHERE server_id = ? AND status = 'active' AND id != ? AND
		((start_time <= ? AND end_time >= ?) OR 
		(start_time <= ? AND end_time >= ?) OR 
		(start_time >= ? AND end_time <= ?))`,
		updatedReservation.ServerID, id,
		updatedReservation.StartTime, updatedReservation.StartTime,
		updatedReservation.EndTime, updatedReservation.EndTime,
		updatedReservation.StartTime, updatedReservation.EndTime,
	).Scan(&count)

	if err != nil {
		slog.Error("Failed to check for overlapping reservations", "server_id", updatedReservation.ServerID, "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for overlapping reservations"})
		return
	}

	if count > 0 {
		slog.Warn("Reservation update failed - overlapping reservation exists", "server_id", updatedReservation.ServerID, "reservation_id", id, "user_id", userID, "overlapping_count", count)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is already reserved for this time period"})
		return
	}

	// Update reservation
	_, err = database.GetDB().Exec(`
		UPDATE reservations 
		SET server_id = ?, server_name = ?, start_time = ?, end_time = ? 
		WHERE id = ? AND user_id = ?`,
		updatedReservation.ServerID, serverName, updatedReservation.StartTime, updatedReservation.EndTime, id, userID,
	)
	if err != nil {
		slog.Error("Failed to update reservation in database", "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reservation"})
		return
	}

	// If server changed, update server status
	if updatedReservation.ServerID != currentReservation.ServerID {
		// Set old server back to available if no other active reservations
		var oldServerReservationCount int
		err = database.GetDB().QueryRow("SELECT COUNT(*) FROM reservations WHERE server_id = ? AND status = 'active'", currentReservation.ServerID).Scan(&oldServerReservationCount)
		if err == nil && oldServerReservationCount == 0 {
			database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "available", currentReservation.ServerID)
		}

		// Set new server to reserved
		_, err = database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "reserved", updatedReservation.ServerID)
		if err != nil {
			slog.Error("Failed to update new server status after reservation update", "server_id", updatedReservation.ServerID, "reservation_id", id, "user_id", userID, "error", err)
		}
	}

	// Return updated reservation
	updatedReservation.ID = id
	updatedReservation.UserID = userID.(int64)
	updatedReservation.Username = username.(string)
	updatedReservation.ServerName = serverName
	updatedReservation.Status = currentReservation.Status

	slog.Info("Reservation updated successfully", "reservation_id", id, "server_id", updatedReservation.ServerID, "user_id", userID, "username", username, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, updatedReservation)
}

// CancelReservation handles reservation cancellation
func CancelReservation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		slog.Warn("Reservation cancellation failed - user not authenticated", "reservation_id", id, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		slog.Warn("Reservation cancellation failed - user role not in context", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not authenticated"})
		return
	}

	// Get reservation details - root can cancel any, user only their own
	var serverID int64
	var status string
	var query string
	var args []interface{}

	if role == "root" {
		query = `SELECT server_id, status FROM reservations WHERE id = ?`
		args = []interface{}{id}
		slog.Info("Reservation cancellation attempt (root access)", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
	} else {
		query = `SELECT server_id, status FROM reservations WHERE id = ? AND user_id = ?`
		args = []interface{}{id, userID}
		slog.Info("Reservation cancellation attempt (user access)", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
	}

	err = database.GetDB().QueryRow(query, args...,
	).Scan(&serverID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Reservation cancellation failed - reservation not found", "reservation_id", id, "user_id", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}
		slog.Error("Failed to fetch reservation for cancellation", "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservation"})
		return
	}

	if status != "active" {
		slog.Warn("Reservation cancellation failed - reservation not active", "reservation_id", id, "status", status, "user_id", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation is not active"})
		return
	}

	// Update reservation status
	_, err = database.GetDB().Exec(`
		UPDATE reservations SET status = ? WHERE id = ?`,
		"cancelled", id,
	)
	if err != nil {
		slog.Error("Failed to cancel reservation in database", "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel reservation"})
		return
	}

	// Update server status
	_, err = database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "available", serverID)
	if err != nil {
		slog.Error("Failed to update server status after cancellation", "server_id", serverID, "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server status"})
		return
	}

	slog.Info("Reservation cancelled successfully", "reservation_id", id, "server_id", serverID, "user_id", userID, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Reservation cancelled successfully"})
}

// DeleteReservation permanently deletes a reservation (root only)
func DeleteReservation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		slog.Warn("Reservation deletion failed - user not authenticated", "reservation_id", id, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, exists := c.Get("role")
	if !exists {
		slog.Warn("Reservation deletion failed - user role not in context", "reservation_id", id, "user_id", userID, "client_ip", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not authenticated"})
		return
	}

	// Only root users can permanently delete reservations
	if role != "root" {
		slog.Warn("Reservation deletion failed - insufficient privileges", "reservation_id", id, "user_id", userID, "role", role, "client_ip", c.ClientIP())
		c.JSON(http.StatusForbidden, gin.H{"error": "Only root users can permanently delete reservations"})
		return
	}

	// Get reservation details before deletion
	var serverID int64
	var status string
	err = database.GetDB().QueryRow(`
		SELECT server_id, status FROM reservations WHERE id = ?`,
		id,
	).Scan(&serverID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("Reservation deletion failed - reservation not found", "reservation_id", id, "user_id", userID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
			return
		}
		slog.Error("Failed to fetch reservation for deletion", "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservation"})
		return
	}

	// Permanently delete the reservation
	_, err = database.GetDB().Exec(`DELETE FROM reservations WHERE id = ?`, id)
	if err != nil {
		slog.Error("Failed to delete reservation from database", "reservation_id", id, "user_id", userID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reservation"})
		return
	}

	// If the reservation was active, update server status to available
	if status == "active" {
		// Check if there are any other active reservations for this server
		var activeCount int
		err = database.GetDB().QueryRow(`
			SELECT COUNT(*) FROM reservations 
			WHERE server_id = ? AND status = 'active'`,
			serverID,
		).Scan(&activeCount)

		if err == nil && activeCount == 0 {
			// No other active reservations, set server to available
			_, err = database.GetDB().Exec("UPDATE servers SET status = ? WHERE id = ?", "available", serverID)
			if err != nil {
				slog.Error("Failed to update server status after reservation deletion", "server_id", serverID, "reservation_id", id, "user_id", userID, "error", err)
			}
		}
	}

	slog.Info("Reservation deleted permanently", "reservation_id", id, "server_id", serverID, "status", status, "user_id", userID, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Reservation deleted permanently"})
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
