package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/config"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"github.com/rusik69/serverscheduler/internal/services"
	"github.com/rusik69/serverscheduler/internal/templates"
)

func render(c *gin.Context, name string, data interface{}) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_ = templates.Execute(c.Writer, name, data)
}

func baseData(c *gin.Context, user services.UserService, cfg config.Config, title, navActive string) templates.BaseData {
	theme, _ := c.Cookie("theme")
	if theme != "dark" && theme != "light" {
		theme = "light"
	}
	bd := templates.BaseData{
		Title:      title,
		Theme:      theme,
		NavActive:  navActive,
		CurrentPath: c.Request.URL.RequestURI(),
	}
	username, ok := middleware.GetCurrentUser(c)
	if !ok {
		// Public pages (e.g. /servers) skip AuthMiddleware; check session cookie directly
		sessionID, _ := c.Cookie("session_id")
		if sessionID != "" {
			if session, exists := middleware.GetSessionStore().GetSession(sessionID); exists {
				username = session.Username
				ok = true
			}
		}
	}
	if !ok {
		return bd
	}
	bd.IsAuthenticated = true
	bd.Username = username
	bd.IsAdmin = username == cfg.AdminUsername
	if !bd.IsAdmin {
		u, _ := user.GetByUsername(c.Request.Context(), username)
		if u != nil && u.Role == "admin" {
			bd.IsAdmin = true
		}
	}
	return bd
}

func isAdmin(c *gin.Context, user services.UserService, cfg config.Config) bool {
	username, ok := middleware.GetCurrentUser(c)
	if !ok {
		return false
	}
	if username == cfg.AdminUsername {
		return true
	}
	u, err := user.GetByUsername(c.Request.Context(), username)
	if err != nil || u == nil {
		return false
	}
	return u.Role == "admin"
}
