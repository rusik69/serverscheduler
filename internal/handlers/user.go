package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/config"
	"github.com/rusik69/serverscheduler/internal/logger"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"github.com/rusik69/serverscheduler/internal/models"
	"github.com/rusik69/serverscheduler/internal/services"
	"github.com/rusik69/serverscheduler/internal/templates"
)

// UserHandler handles user/profile endpoints
type UserHandler struct {
	user        services.UserService
	reservation services.ReservationService
	server      services.ServerService
	ssh         services.SSHService
	config      config.Config
}

// NewUserHandler creates a UserHandler
func NewUserHandler(user services.UserService, res services.ReservationService, srv services.ServerService, ssh services.SSHService, cfg config.Config) *UserHandler {
	return &UserHandler{user: user, reservation: res, server: srv, ssh: ssh, config: cfg}
}

// ProfileData for template
type ProfileData struct {
	Username     string
	Role         string
	SSHPublicKey string
}

// ProfilePage renders the profile page
func (h *UserHandler) ProfilePage(c *gin.Context) {
	username, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	logger.FromContext(c.Request.Context()).Debug("profile page load", "username", username)
	bd := baseData(c, h.user, h.config, "Profile", "profile")
	profile := ProfileData{Username: username, Role: "admin"}
	if username != h.config.AdminUsername {
		u, err := h.user.GetByUsername(c.Request.Context(), username)
		if err != nil || u == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		profile = ProfileData{Username: u.Username, Role: u.Role, SSHPublicKey: u.SSHPublicKey}
	}
	data := struct {
		templates.BaseData
		Profile ProfileData
		Error   string
		Success string
	}{BaseData: bd, Profile: profile, Error: c.Query("error"), Success: c.Query("success")}
	render(c, "profile", data)
}

// UpdateSSHKey handles form POST
func (h *UserHandler) UpdateSSHKey(c *gin.Context) {
	username, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if username == h.config.AdminUsername {
		c.Redirect(http.StatusFound, "/profile?error=admin+cannot+set+SSH+key")
		return
	}
	sshKey := c.PostForm("ssh_public_key")
	if err := h.user.UpdateSSHKey(c.Request.Context(), username, sshKey); err != nil {
		logger.FromContext(c.Request.Context()).Error("update SSH key failed", "username", username, "error", err)
		c.Redirect(http.StatusFound, "/profile?error="+err.Error())
		return
	}
	logger.FromContext(c.Request.Context()).Info("SSH key updated", "username", username)
	c.Redirect(http.StatusFound, "/profile?success=Saved")
}

// UsersPage renders the users list (admin only)
func (h *UserHandler) UsersPage(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/servers?error=admin+required")
		return
	}
	logger.FromContext(c.Request.Context()).Debug("users page load")
	list, err := h.user.List(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	bd := baseData(c, h.user, h.config, "Users", "users")
	data := struct {
		templates.BaseData
		Users         interface{}
		AdminUsername string
		Error         string
		Success       string
	}{BaseData: bd, Users: list, AdminUsername: h.config.AdminUsername, Error: c.Query("error"), Success: c.Query("success")}
	render(c, "users", data)
}

// DeleteUser handles form POST (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/users?error=admin+required")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/users?error=invalid+id")
		return
	}
	u, err := h.user.GetByID(c.Request.Context(), id)
	if err != nil || u == nil {
		c.Redirect(http.StatusFound, "/users?error=user+not+found")
		return
	}
	currentUser, _ := middleware.GetCurrentUser(c)
	if u.Username == h.config.AdminUsername {
		c.Redirect(http.StatusFound, "/users?error=cannot+delete+admin+user")
		return
	}
	if u.Username == currentUser {
		c.Redirect(http.StatusFound, "/users?error=cannot+delete+yourself")
		return
	}
	// Revoke SSH access for active reservations, then delete all reservations for this user
	userID := &u.ID
	reservations, _ := h.reservation.List(c.Request.Context(), userID)
	for _, r := range reservations {
		if r.Status == "active" {
			h.revokeAccess(c.Request.Context(), &r.Reservation)
		}
	}
	_ = h.reservation.DeleteByUserID(c.Request.Context(), u.ID)
	if err := h.user.Delete(c.Request.Context(), id); err != nil {
		logger.FromContext(c.Request.Context()).Error("delete user failed", "user_id", id, "username", u.Username, "error", err)
		c.Redirect(http.StatusFound, "/users?error="+err.Error())
		return
	}
	logger.FromContext(c.Request.Context()).Info("user deleted", "user_id", id, "username", u.Username)
	c.Redirect(http.StatusFound, "/users?success=User+removed")
}

func (h *UserHandler) revokeAccess(ctx context.Context, r *models.Reservation) {
	usr, _ := h.user.GetByID(ctx, r.UserID)
	if usr == nil || usr.SSHPublicKey == "" {
		return
	}
	srv, _ := h.server.Get(ctx, r.ServerID)
	if srv == nil {
		return
	}
	_ = h.ssh.RemoveKey(ctx, srv.Hostname, srv.Port, srv.SSHUser, srv.SSHPrivateKey, usr.SSHPublicKey)
}
