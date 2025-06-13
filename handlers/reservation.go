package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/database"
	"github.com/rusik69/serverscheduler/models"
)

// CreateReservation handles reservation creation
func CreateReservation(c *gin.Context) {
	var req models.ReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if server exists and is available
	var serverStatus string
	err := database.DB.QueryRow("SELECT status FROM servers WHERE id = ?", req.ServerID).Scan(&serverStatus)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if serverStatus != "available" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is not available"})
		return
	}

	// Check for overlapping reservations
	var count int
	err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM reservations 
		WHERE server_id = ? AND status = 'active' 
		AND ((start_time <= ? AND end_time >= ?) OR (start_time <= ? AND end_time >= ?))`,
		req.ServerID, req.EndTime, req.StartTime, req.EndTime, req.StartTime,
	).Scan(&count)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server is already reserved for this time period"})
		return
	}

	userID, _ := c.Get("userID")
	result, err := database.DB.Exec(`
		INSERT INTO reservations (server_id, user_id, start_time, end_time, status)
		VALUES (?, ?, ?, ?, ?)`,
		req.ServerID, userID, req.StartTime, req.EndTime, "active",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	reservationID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": reservationID})
}

// ListReservations handles listing all reservations
func ListReservations(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	var rows *sql.Rows
	var err error

	if role == "admin" {
		rows, err = database.DB.Query(`
			SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, r.created_at, r.updated_at
			FROM reservations r
			ORDER BY r.start_time DESC
		`)
	} else {
		rows, err = database.DB.Query(`
			SELECT r.id, r.server_id, r.user_id, r.start_time, r.end_time, r.status, r.created_at, r.updated_at
			FROM reservations r
			WHERE r.user_id = ?
			ORDER BY r.start_time DESC
		`, userID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(
			&reservation.ID, &reservation.ServerID, &reservation.UserID,
			&reservation.StartTime, &reservation.EndTime, &reservation.Status,
			&reservation.CreatedAt, &reservation.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan reservation"})
			return
		}
		reservations = append(reservations, reservation)
	}

	c.JSON(http.StatusOK, reservations)
}

// CancelReservation handles reservation cancellation
func CancelReservation(c *gin.Context) {
	reservationID := c.Param("id")
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	var result sql.Result
	var err error

	if role == "admin" {
		result, err = database.DB.Exec(
			"UPDATE reservations SET status = 'cancelled', updated_at = ? WHERE id = ?",
			time.Now(), reservationID,
		)
	} else {
		result, err = database.DB.Exec(
			"UPDATE reservations SET status = 'cancelled', updated_at = ? WHERE id = ? AND user_id = ?",
			time.Now(), reservationID, userID,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel reservation"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation cancelled successfully"})
}
