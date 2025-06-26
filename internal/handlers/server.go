package handlers

import (
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/database"
	"github.com/rusik69/serverscheduler/internal/models"
)

// validateIPAddress validates if the given string is a valid IP address
func validateIPAddress(ip string) bool {
	// Allow empty IP address
	if strings.TrimSpace(ip) == "" {
		return true
	}

	// Parse IP address
	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	return parsedIP != nil
}

// CreateServer handles server creation
func CreateServer(c *gin.Context) {
	var server models.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		slog.Warn("Invalid server creation request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate IP address format
	if !validateIPAddress(server.IPAddress) {
		slog.Warn("Server creation failed - invalid IP address format", "ip_address", server.IPAddress, "server_name", server.Name, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IP address format"})
		return
	}

	slog.Info("Server creation attempt", "server_name", server.Name, "client_ip", c.ClientIP())

	result, err := database.GetDB().Exec(
		"INSERT INTO servers (name, status, ip_address, username, password) VALUES (?, ?, ?, ?, ?)",
		server.Name, "available", server.IPAddress, server.Username, server.Password,
	)
	if err != nil {
		slog.Error("Failed to create server in database", "server_name", server.Name, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create server"})
		return
	}

	id, _ := result.LastInsertId()
	server.ID = id
	server.Status = "available"

	slog.Info("Server created successfully", "server_id", server.ID, "server_name", server.Name, "ip_address", server.IPAddress, "client_ip", c.ClientIP())
	c.JSON(http.StatusCreated, server)
}

// GetServers returns all servers
func GetServers(c *gin.Context) {
	slog.Info("Servers list requested", "client_ip", c.ClientIP())

	// Get user role from context
	role, exists := c.Get("role")
	isRoot := exists && role == "root"

	rows, err := database.GetDB().Query("SELECT id, name, status, COALESCE(ip_address, '') as ip_address, COALESCE(username, '') as username, COALESCE(password, '') as password FROM servers")
	if err != nil {
		slog.Error("Failed to fetch servers from database", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch servers"})
		return
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var server models.Server
		if err := rows.Scan(&server.ID, &server.Name, &server.Status, &server.IPAddress, &server.Username, &server.Password); err != nil {
			slog.Error("Failed to scan server row", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan server"})
			return
		}

		// Hide password from non-root users
		if !isRoot {
			server.Password = ""
		}

		servers = append(servers, server)
	}

	slog.Info("Servers list returned", "count", len(servers), "client_ip", c.ClientIP(), "show_passwords", isRoot)
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

	// Get user role from context
	role, exists := c.Get("role")
	isRoot := exists && role == "root"

	var server models.Server
	err = database.GetDB().QueryRow(
		"SELECT id, name, status, COALESCE(ip_address, '') as ip_address, COALESCE(username, '') as username, COALESCE(password, '') as password FROM servers WHERE id = ?",
		id,
	).Scan(&server.ID, &server.Name, &server.Status, &server.IPAddress, &server.Username, &server.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch server"})
		return
	}

	// Hide password from non-root users
	if !isRoot {
		server.Password = ""
	}

	slog.Info("Server details returned", "server_id", id, "client_ip", c.ClientIP(), "show_password", isRoot)
	c.JSON(http.StatusOK, server)
}

// UpdateServer handles server updates
func UpdateServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Warn("Invalid server ID for update", "id_param", idStr, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var server models.Server
	if err := c.ShouldBindJSON(&server); err != nil {
		slog.Warn("Invalid server update request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate IP address format
	if !validateIPAddress(server.IPAddress) {
		slog.Warn("Server update failed - invalid IP address format", "server_id", id, "ip_address", server.IPAddress, "server_name", server.Name, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IP address format"})
		return
	}

	slog.Info("Server update attempt", "server_id", id, "server_name", server.Name, "ip_address", server.IPAddress, "client_ip", c.ClientIP())

	_, err = database.GetDB().Exec(
		"UPDATE servers SET name = ?, status = ?, ip_address = ?, username = ?, password = ? WHERE id = ?",
		server.Name, server.Status, server.IPAddress, server.Username, server.Password, id,
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
		slog.Warn("Invalid server ID for deletion", "id_param", idStr, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	slog.Info("Server deletion attempt", "server_id", id, "client_ip", c.ClientIP())

	_, err = database.GetDB().Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		slog.Error("Failed to delete server from database", "server_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete server"})
		return
	}

	slog.Info("Server deleted successfully", "server_id", id, "client_ip", c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Server deleted successfully"})
}
