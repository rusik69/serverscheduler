package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
)

// CreateServer handles server creation
func CreateServer(c *gin.Context) {
	var server models.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := database.GetDB().Exec(
		"INSERT INTO servers (name, status) VALUES (?, ?)",
		server.Name, "available",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	id, _ := result.LastInsertId()
	server.ID = id
	server.Status = "available"

	c.JSON(http.StatusCreated, server)
}

// GetServers returns all servers
func GetServers(c *gin.Context) {
	rows, err := database.GetDB().Query("SELECT id, name, status FROM servers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var server models.Server
		if err := rows.Scan(&server.ID, &server.Name, &server.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan server"})
			return
		}
		servers = append(servers, server)
	}

	c.JSON(http.StatusOK, gin.H{"servers": servers})
}

// GetServer returns a specific server
func GetServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var server models.Server
	err = database.GetDB().QueryRow(
		"SELECT id, name, status FROM servers WHERE id = ?",
		id,
	).Scan(&server.ID, &server.Name, &server.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch server"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// UpdateServer handles server updates
func UpdateServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var server models.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = database.GetDB().Exec(
		"UPDATE servers SET name = ?, status = ? WHERE id = ?",
		server.Name, server.Status, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update server"})
		return
	}

	server.ID = id
	c.JSON(http.StatusOK, server)
}

// DeleteServer handles server deletion
func DeleteServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	_, err = database.GetDB().Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Server deleted successfully"})
}
