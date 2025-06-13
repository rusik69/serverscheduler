package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/database"
	"github.com/rusik69/serverscheduler/models"
)

// CreateServer handles server creation (root only)
func CreateServer(c *gin.Context) {
	// Check if user is root
	role, exists := c.Get("role")
	if !exists || role != "root" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only root user can create servers"})
		return
	}

	var req models.Server
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server name cannot be empty"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO servers (name, description, status) VALUES (?, ?, ?)",
		req.Name, req.Description, "available",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	serverID, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": serverID})
}

// ListServers handles listing all servers
func ListServers(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, name, description, status, created_at, updated_at FROM servers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var server models.Server
		err := rows.Scan(&server.ID, &server.Name, &server.Description, &server.Status, &server.CreatedAt, &server.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan server"})
			return
		}
		servers = append(servers, server)
	}

	c.JSON(http.StatusOK, gin.H{"servers": servers})
}
