package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/config"
	"github.com/rusik69/serverscheduler/internal/logger"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"github.com/rusik69/serverscheduler/internal/services"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles auth endpoints
type AuthHandler struct {
	user   services.UserService
	config config.Config
}

// NewAuthHandler creates an AuthHandler
func NewAuthHandler(user services.UserService, cfg config.Config) *AuthHandler {
	return &AuthHandler{user: user, config: cfg}
}

// LoginPage renders the login form
func (h *AuthHandler) LoginPage(c *gin.Context) {
	if sessionID, _ := c.Cookie("session_id"); sessionID != "" {
		if _, exists := middleware.GetSessionStore().GetSession(sessionID); exists {
			c.Redirect(http.StatusFound, "/reservations")
			return
		}
	}
	bd := baseData(c, h.user, h.config, "Login", "")
	bd.Error = c.Query("error")
	render(c, "login", bd)
}

// Login handles form POST
func (h *AuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		logger.FromContext(c.Request.Context()).Warn("login failed", "username", username, "error", "username and password required")
		c.Redirect(http.StatusFound, "/login?error=username+and+password+required")
		return
	}

	if username == h.config.AdminUsername {
		if password != h.config.AdminPassword {
			logger.FromContext(c.Request.Context()).Warn("login failed", "username", username, "error", "invalid credentials")
			c.Redirect(http.StatusFound, "/login?error=Invalid+credentials")
			return
		}
		sessionID := middleware.GenerateSessionID()
		middleware.GetSessionStore().SetSession(sessionID, username)
		c.SetCookie("session_id", sessionID, int(24*time.Hour.Seconds()), "/", "", false, false)
		logger.FromContext(c.Request.Context()).Info("login success", "username", username, "role", "admin")
		c.Redirect(http.StatusFound, "/reservations")
		return
	}

	u, err := h.user.GetByUsername(c.Request.Context(), username)
	if err != nil || u == nil {
		logger.FromContext(c.Request.Context()).Warn("login failed", "username", username, "error", "invalid credentials")
		c.Redirect(http.StatusFound, "/login?error=Invalid+credentials")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		logger.FromContext(c.Request.Context()).Warn("login failed", "username", username, "error", "invalid credentials")
		c.Redirect(http.StatusFound, "/login?error=Invalid+credentials")
		return
	}

	sessionID := middleware.GenerateSessionID()
	middleware.GetSessionStore().SetSession(sessionID, username)
	c.SetCookie("session_id", sessionID, int(24*time.Hour.Seconds()), "/", "", false, false)
	logger.FromContext(c.Request.Context()).Info("login success", "username", username)
	c.Redirect(http.StatusFound, "/reservations")
}

// RegisterPage renders the register form
func (h *AuthHandler) RegisterPage(c *gin.Context) {
	bd := baseData(c, h.user, h.config, "Register", "")
	bd.Error = c.Query("error")
	render(c, "register", bd)
}

// Register handles form POST
func (h *AuthHandler) Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	sshKey := c.PostForm("ssh_public_key")
	if username == "" || password == "" || sshKey == "" {
		c.Redirect(http.StatusFound, "/register?error=username+password+and+SSH+key+required")
		return
	}
	if len(password) < 6 {
		c.Redirect(http.StatusFound, "/register?error=password+must+be+at+least+6+characters")
		return
	}
	if username == h.config.AdminUsername {
		c.Redirect(http.StatusFound, "/register?error=username+not+allowed")
		return
	}

	_, err := h.user.CreateWithSSHKey(c.Request.Context(), username, password, sshKey)
	if err != nil {
		logger.FromContext(c.Request.Context()).Warn("register failed", "username", username, "error", err)
		c.Redirect(http.StatusFound, "/register?error=username+already+exists")
		return
	}
	logger.FromContext(c.Request.Context()).Info("register success", "username", username)
	c.Redirect(http.StatusFound, "/login")
}

// RegisterUser handles form POST (admin only) - creates regular user
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/users?error=Admin+required")
		return
	}
	username := c.PostForm("username")
	password := c.PostForm("password")
	sshKey := c.PostForm("ssh_public_key")
	if username == "" || password == "" {
		c.Redirect(http.StatusFound, "/users?error=username+and+password+required")
		return
	}
	if len(password) < 6 {
		c.Redirect(http.StatusFound, "/users?error=password+must+be+at+least+6+characters")
		return
	}
	if username == h.config.AdminUsername {
		c.Redirect(http.StatusFound, "/users?error=username+not+allowed")
		return
	}
	_, err := h.user.CreateWithSSHKey(c.Request.Context(), username, password, sshKey)
	if err != nil {
		logger.FromContext(c.Request.Context()).Warn("register user failed", "username", username, "error", err)
		c.Redirect(http.StatusFound, "/users?error=username+already+exists")
		return
	}
	logger.FromContext(c.Request.Context()).Info("register user success", "username", username)
	c.Redirect(http.StatusFound, "/users?success=User+registered")
}

// RegisterAdmin handles form POST (admin only)
func (h *AuthHandler) RegisterAdmin(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/users?error=Admin+required")
		return
	}
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.Redirect(http.StatusFound, "/users?error=username+and+password+required")
		return
	}
	if len(password) < 6 {
		c.Redirect(http.StatusFound, "/users?error=password+must+be+at+least+6+characters")
		return
	}
	_, err := h.user.Create(c.Request.Context(), username, password, "admin")
	if err != nil {
		logger.FromContext(c.Request.Context()).Warn("register admin failed", "username", username, "error", err)
		c.Redirect(http.StatusFound, "/users?error=username+already+exists")
		return
	}
	logger.FromContext(c.Request.Context()).Info("register admin success", "username", username)
	c.Redirect(http.StatusFound, "/users?success=Admin+registered")
}

// Logout handles form POST
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, _ := c.Cookie("session_id")
	if sessionID != "" {
		middleware.GetSessionStore().DeleteSession(sessionID)
	}
	c.SetCookie("session_id", "", -1, "/", "", false, false)
	logger.FromContext(c.Request.Context()).Info("logout")
	c.Redirect(http.StatusFound, "/servers")
}
