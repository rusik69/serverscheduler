package handlers

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/config"
	"github.com/rusik69/serverscheduler/internal/logger"
	"github.com/rusik69/serverscheduler/internal/models"
	"github.com/rusik69/serverscheduler/internal/services"
	"github.com/rusik69/serverscheduler/internal/templates"
)

// ServerWithUsers extends Server with users who have access (pending/active reservations)
type ServerWithUsers struct {
	models.Server
	Users               []string
	CurrentReservation  *models.ReservationWithDetails
}

// ServerHandler handles server endpoints
type ServerHandler struct {
	server     services.ServerService
	reservation services.ReservationService
	ssh        services.SSHService
	user       services.UserService
	config     config.Config
}

// NewServerHandler creates a ServerHandler
func NewServerHandler(server services.ServerService, res services.ReservationService, ssh services.SSHService, user services.UserService, cfg config.Config) *ServerHandler {
	return &ServerHandler{server: server, reservation: res, ssh: ssh, user: user, config: cfg}
}

// ServersPage renders the servers list
func (h *ServerHandler) ServersPage(c *gin.Context) {
	list, err := h.server.List(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	usersByServer, _ := h.reservation.GetUsersByServer(c.Request.Context())
	currentByServer, _ := h.reservation.GetCurrentByServer(c.Request.Context())
	serversWithUsers := make([]ServerWithUsers, len(list))
	for i, s := range list {
		serversWithUsers[i] = ServerWithUsers{
			Server:              s,
			Users:               usersByServer[s.ID],
			CurrentReservation:  currentByServer[s.ID],
		}
	}
	bd := baseData(c, h.user, h.config, "Servers", "servers")
	data := struct {
		templates.BaseData
		Servers []ServerWithUsers
		Error   string
		Success string
	}{BaseData: bd, Servers: serversWithUsers, Error: c.Query("error"), Success: c.Query("success")}
	render(c, "servers", data)
}

// AddServer handles form POST
func (h *ServerHandler) AddServer(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/servers?error=admin+required")
		return
	}
	name := c.PostForm("name")
	hostname := c.PostForm("hostname")
	sshUser := c.PostForm("ssh_user")
	sshKey := c.PostForm("ssh_private_key")
	if name == "" || hostname == "" || sshUser == "" || sshKey == "" {
		c.Redirect(http.StatusFound, "/servers?error=name+hostname+ssh_user+and+ssh_private_key+required")
		return
	}
	port := 22
	if p := c.PostForm("port"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	srv := &models.Server{
		Name:          name,
		Hostname:      hostname,
		Port:          port,
		SSHUser:       sshUser,
		SSHPrivateKey: sshKey,
		Description:   c.PostForm("description"),
	}
	if err := h.ssh.TestConnection(c.Request.Context(), hostname, port, sshUser, sshKey); err != nil {
		logger.FromContext(c.Request.Context()).Error("add server SSH test failed", "name", name, "error", err)
		c.Redirect(http.StatusFound, "/servers?error="+url.QueryEscape("SSH connection failed: "+err.Error()))
		return
	}
	created, err := h.server.Create(c.Request.Context(), srv)
	if err != nil {
		logger.FromContext(c.Request.Context()).Error("add server failed", "name", name, "error", err)
		c.Redirect(http.StatusFound, "/servers?error="+url.QueryEscape(err.Error()))
		return
	}
	logger.FromContext(c.Request.Context()).Info("server added", "server_id", created.ID, "name", name)
	c.Redirect(http.StatusFound, "/servers")
}

// TestServer handles form POST - tests SSH connection
func (h *ServerHandler) TestServer(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/servers?error=admin+required")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/servers?error=invalid+id")
		return
	}
	srv, err := h.server.Get(c.Request.Context(), id)
	if err != nil || srv == nil {
		c.Redirect(http.StatusFound, "/servers?error=server+not+found")
		return
	}
	if err := h.ssh.TestConnection(c.Request.Context(), srv.Hostname, srv.Port, srv.SSHUser, srv.SSHPrivateKey); err != nil {
		logger.FromContext(c.Request.Context()).Error("server SSH test failed", "server_id", id, "name", srv.Name, "error", err)
		c.Redirect(http.StatusFound, "/servers?error="+url.QueryEscape("SSH test failed: "+err.Error()))
		return
	}
	logger.FromContext(c.Request.Context()).Info("server SSH test success", "server_id", id, "name", srv.Name)
	c.Redirect(http.StatusFound, "/servers?success=SSH+connection+OK")
}

// DeleteServer handles form POST
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/servers?error=admin+required")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/servers?error=invalid+id")
		return
	}
	if err := h.server.Delete(c.Request.Context(), id); err != nil {
		logger.FromContext(c.Request.Context()).Error("delete server failed", "server_id", id, "error", err)
		c.Redirect(http.StatusFound, "/servers?error="+err.Error())
		return
	}
	logger.FromContext(c.Request.Context()).Info("server deleted", "server_id", id)
	c.Redirect(http.StatusFound, "/servers")
}
