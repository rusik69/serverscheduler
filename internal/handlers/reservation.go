package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rusik69/serverscheduler/internal/config"
	"github.com/rusik69/serverscheduler/internal/logger"
	"github.com/rusik69/serverscheduler/internal/middleware"
	"github.com/rusik69/serverscheduler/internal/models"
	"github.com/rusik69/serverscheduler/internal/services"
	"github.com/rusik69/serverscheduler/internal/templates"
)

// ReservationHandler handles reservation endpoints
type ReservationHandler struct {
	reservation services.ReservationService
	server      services.ServerService
	user        services.UserService
	ssh         services.SSHService
	config      config.Config
}

// NewReservationHandler creates a ReservationHandler
func NewReservationHandler(res services.ReservationService, srv services.ServerService, user services.UserService, ssh services.SSHService, cfg config.Config) *ReservationHandler {
	return &ReservationHandler{reservation: res, server: srv, user: user, ssh: ssh, config: cfg}
}

// reservationDataItem is the JSON shape for /reservations/data
type reservationDataItem struct {
	ID              int64  `json:"id"`
	ServerName      string `json:"server_name"`
	Username        string `json:"username"`
	StartFormatted  string `json:"start_formatted"`
	EndFormatted    string `json:"end_formatted"`
	StartUTC        string `json:"start_utc"`
	EndUTC          string `json:"end_utc"`
	Status          string `json:"status"`
	CanCancel       bool   `json:"can_cancel"`
}

// ReservationsData returns reservations as JSON for polling
func (h *ReservationHandler) ReservationsData(c *gin.Context) {
	username, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u, _ := h.user.GetByUsername(c.Request.Context(), username)
	isAdmin := username == h.config.AdminUsername || (u != nil && u.Role == "admin")

	var userID *int64
	if !isAdmin && u != nil {
		userID = &u.ID
	}
	reservations, err := h.reservation.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]reservationDataItem, len(reservations))
	for i, r := range reservations {
		startFmt := "-"
		endFmt := "-"
		startISO := ""
		endISO := ""
		if !r.StartTime.IsZero() {
			startFmt = r.StartTime.UTC().Format("2006-01-02 15:04 UTC")
			startISO = r.StartTime.UTC().Format(time.RFC3339)
		}
		if !r.EndTime.IsZero() {
			endFmt = r.EndTime.UTC().Format("2006-01-02 15:04 UTC")
			endISO = r.EndTime.UTC().Format(time.RFC3339)
		}
		items[i] = reservationDataItem{
			ID:             r.ID,
			ServerName:     r.ServerName,
			Username:       r.Username,
			StartFormatted: startFmt,
			EndFormatted:   endFmt,
			StartUTC:       startISO,
			EndUTC:         endISO,
			Status:        r.Status,
			CanCancel:     r.Status == "pending" || r.Status == "active",
		}
	}
	c.JSON(http.StatusOK, items)
}

// ReservationsPage renders the reservations list
func (h *ReservationHandler) ReservationsPage(c *gin.Context) {
	username, _ := middleware.GetCurrentUser(c)
	u, _ := h.user.GetByUsername(c.Request.Context(), username)
	isAdmin := username == h.config.AdminUsername || (u != nil && u.Role == "admin")
	canCreate := !isAdmin

	var userID *int64
	if !isAdmin && u != nil {
		userID = &u.ID
	}
	reservations, err := h.reservation.List(c.Request.Context(), userID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	servers, _ := h.server.List(c.Request.Context())
	var users []models.UserPublic
	if isAdmin {
		allUsers, _ := h.user.List(c.Request.Context())
		for _, uu := range allUsers {
			if uu.Role == "user" {
				users = append(users, uu)
			}
		}
	}
	bd := baseData(c, h.user, h.config, "Reservations", "reservations")
	data := struct {
		templates.BaseData
		Reservations []models.ReservationWithDetails
		Servers      []models.Server
		Users        []models.UserPublic
		CanCreate    bool
		IsAdmin      bool
		Error        string
	}{BaseData: bd, Reservations: reservations, Servers: servers, Users: users, CanCreate: canCreate, IsAdmin: isAdmin, Error: c.Query("error")}
	render(c, "reservations", data)
}

// AddReservation handles form POST
func (h *ReservationHandler) AddReservation(c *gin.Context) {
	username, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if username == h.config.AdminUsername {
		c.Redirect(http.StatusFound, "/reservations?error=admin+cannot+create+reservations")
		return
	}
	u, err := h.user.GetByUsername(c.Request.Context(), username)
	if err != nil || u == nil {
		c.Redirect(http.StatusFound, "/reservations?error=user+not+found")
		return
	}

	serverID, err := strconv.ParseInt(c.PostForm("server_id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+server")
		return
	}
	startStr := c.PostForm("start_time")
	endStr := c.PostForm("end_time")
	if startStr == "" || endStr == "" {
		c.Redirect(http.StatusFound, "/reservations?error=start+and+end+time+required")
		return
	}
	start, err := parseDateTimeUTC(startStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+start_time")
		return
	}
	end, err := parseDateTimeUTC(endStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+end_time")
		return
	}
	if !end.After(start) {
		c.Redirect(http.StatusFound, "/reservations?error=end+must+be+after+start")
		return
	}
	if start.Before(time.Now().UTC()) {
		c.Redirect(http.StatusFound, "/reservations?error=start+time+cannot+be+in+the+past")
		return
	}

	r, err := h.reservation.Create(c.Request.Context(), u.ID, serverID, start, end)
	if err != nil {
		if err == services.ErrOverlap {
			logger.FromContext(c.Request.Context()).Warn("reservation create failed", "user_id", u.ID, "server_id", serverID, "error", "overlap")
			c.Redirect(http.StatusFound, "/reservations?error=reservation+overlaps")
			return
		}
		logger.FromContext(c.Request.Context()).Error("reservation create failed", "user_id", u.ID, "server_id", serverID, "error", err)
		c.Redirect(http.StatusFound, "/reservations?error="+err.Error())
		return
	}
	logger.FromContext(c.Request.Context()).Info("reservation created", "reservation_id", r.ID, "user_id", u.ID, "server_id", serverID)
	c.Redirect(http.StatusFound, "/reservations")
}

// CancelReservation handles form POST
func (h *ReservationHandler) CancelReservation(c *gin.Context) {
	username, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+id")
		return
	}
	r, _ := h.reservation.Get(c.Request.Context(), id)
	if r == nil {
		c.Redirect(http.StatusFound, "/reservations?error=reservation+not+found")
		return
	}
	if !isAdmin(c, h.user, h.config) {
		u, _ := h.user.GetByUsername(c.Request.Context(), username)
		if u == nil || u.ID != r.UserID {
			c.Redirect(http.StatusFound, "/reservations?error=reservation+not+found")
			return
		}
	}
	if r.Status == "active" || r.Status == "pending" {
		h.revokeAccess(c.Request.Context(), r)
	}
	if isAdmin(c, h.user, h.config) {
		if err := h.reservation.CancelByAdmin(c.Request.Context(), id); err != nil {
			logger.FromContext(c.Request.Context()).Error("admin cancel reservation failed", "reservation_id", id, "error", err)
			c.Redirect(http.StatusFound, "/reservations?error="+err.Error())
			return
		}
	} else {
		if err := h.reservation.Cancel(c.Request.Context(), id, r.UserID); err != nil {
			if err == services.ErrNotFound {
				c.Redirect(http.StatusFound, "/reservations?error=reservation+not+found")
				return
			}
			logger.FromContext(c.Request.Context()).Error("cancel reservation failed", "reservation_id", id, "user_id", r.UserID, "error", err)
			c.Redirect(http.StatusFound, "/reservations?error="+err.Error())
			return
		}
	}
	logger.FromContext(c.Request.Context()).Info("reservation cancelled", "reservation_id", id, "user_id", r.UserID)
	c.Redirect(http.StatusFound, "/reservations")
}

func (h *ReservationHandler) revokeAccess(ctx context.Context, r *models.Reservation) {
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

// AdminAddReservation handles form POST (admin only) - creates reservation for a user
func (h *ReservationHandler) AdminAddReservation(c *gin.Context) {
	if !isAdmin(c, h.user, h.config) {
		c.Redirect(http.StatusFound, "/reservations?error=admin+required")
		return
	}
	userID, err := strconv.ParseInt(c.PostForm("user_id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+user")
		return
	}
	serverID, err := strconv.ParseInt(c.PostForm("server_id"), 10, 64)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+server")
		return
	}
	startStr := c.PostForm("start_time")
	endStr := c.PostForm("end_time")
	if startStr == "" || endStr == "" {
		c.Redirect(http.StatusFound, "/reservations?error=start+and+end+time+required")
		return
	}
	start, err := parseDateTimeUTC(startStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+start_time")
		return
	}
	end, err := parseDateTimeUTC(endStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/reservations?error=invalid+end_time")
		return
	}
	if !end.After(start) {
		c.Redirect(http.StatusFound, "/reservations?error=end+must+be+after+start")
		return
	}
	if start.Before(time.Now().UTC()) {
		c.Redirect(http.StatusFound, "/reservations?error=start+time+cannot+be+in+the+past")
		return
	}
	r, err := h.reservation.Create(c.Request.Context(), userID, serverID, start, end)
	if err != nil {
		if err == services.ErrOverlap {
			logger.FromContext(c.Request.Context()).Warn("admin reservation create failed", "user_id", userID, "server_id", serverID, "error", "overlap")
			c.Redirect(http.StatusFound, "/reservations?error=reservation+overlaps")
			return
		}
		logger.FromContext(c.Request.Context()).Error("admin reservation create failed", "user_id", userID, "server_id", serverID, "error", err)
		c.Redirect(http.StatusFound, "/reservations?error="+err.Error())
		return
	}
	logger.FromContext(c.Request.Context()).Info("admin reservation created", "reservation_id", r.ID, "user_id", userID, "server_id", serverID)
	c.Redirect(http.StatusFound, "/reservations")
}

// parseDateTimeUTC parses datetime-local value as UTC
func parseDateTimeUTC(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04", "2006-01-02T15:04:05"} {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, errors.New("invalid datetime format")
}
