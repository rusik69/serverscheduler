package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/rusik69/serverscheduler/internal/models"
)

// Scheduler runs background activation/expiration of reservations
type Scheduler struct {
	reservation ReservationService
	server     ServerService
	user       UserService
	ssh        SSHService
	slack      SlackService
	interval   time.Duration
}

// NewScheduler creates a Scheduler
func NewScheduler(res ReservationService, srv ServerService, usr UserService, ssh SSHService, slack SlackService) *Scheduler {
	return &Scheduler{
		reservation: res,
		server:     srv,
		user:       usr,
		ssh:        ssh,
		slack:      slack,
		interval:   60 * time.Second,
	}
}

// Start runs the scheduler loop
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	slog.Debug("scheduler tick start")
	pending, err := s.reservation.GetPendingToActivate(ctx)
	if err != nil {
		slog.Error("scheduler get pending failed", "error", err)
		return
	}
	slog.Debug("scheduler tick", "pending_count", len(pending))
	for _, r := range pending {
		if err := s.activateReservation(ctx, r); err != nil {
			slog.Error("scheduler activate reservation failed", "reservation_id", r.ID, "user_id", r.UserID, "server_id", r.ServerID, "error", err)
		}
	}

	active, err := s.reservation.GetActiveToExpire(ctx)
	if err != nil {
		slog.Error("scheduler get active failed", "error", err)
		return
	}
	slog.Debug("scheduler tick", "active_count", len(active))
	for _, r := range active {
		if err := s.expireReservation(ctx, r); err != nil {
			slog.Error("scheduler expire reservation failed", "reservation_id", r.ID, "user_id", r.UserID, "server_id", r.ServerID, "error", err)
		}
	}
}

func (s *Scheduler) activateReservation(ctx context.Context, r models.Reservation) error {
	usr, err := s.user.GetByID(ctx, r.UserID)
	if err != nil || usr == nil || usr.SSHPublicKey == "" {
		slog.Warn("scheduler user has no SSH key, skipping activation", "reservation_id", r.ID, "user_id", r.UserID)
		return s.reservation.Activate(ctx, r.ID)
	}
	srv, err := s.server.Get(ctx, r.ServerID)
	if err != nil || srv == nil {
		return err
	}
	if err := s.ssh.AddKey(ctx, srv.Hostname, srv.Port, srv.SSHUser, srv.SSHPrivateKey, usr.SSHPublicKey); err != nil {
		return err
	}
	if err := s.reservation.Activate(ctx, r.ID); err != nil {
		return err
	}
	slog.Info("reservation activated", "reservation_id", r.ID, "user_id", r.UserID, "server_id", r.ServerID, "username", usr.Username)
	msg := fmt.Sprintf("Reservation activated: user %s now has SSH access to %s (until %s)", usr.Username, srv.Name, r.EndTime.Format(time.RFC3339))
	if err := s.slack.Notify(ctx, msg); err != nil {
		slog.Warn("slack notify failed", "reservation_id", r.ID, "error", err)
	}
	return nil
}

func (s *Scheduler) expireReservation(ctx context.Context, r models.Reservation) error {
	usr, err := s.user.GetByID(ctx, r.UserID)
	if err != nil || usr == nil {
		return s.reservation.Expire(ctx, r.ID)
	}
	srv, err := s.server.Get(ctx, r.ServerID)
	if err != nil || srv == nil {
		return err
	}
	if usr.SSHPublicKey != "" {
		if err := s.ssh.RemoveKey(ctx, srv.Hostname, srv.Port, srv.SSHUser, srv.SSHPrivateKey, usr.SSHPublicKey); err != nil {
			slog.Warn("scheduler remove key failed", "reservation_id", r.ID, "user_id", r.UserID, "server_id", r.ServerID, "error", err)
		}
	}
	if err := s.reservation.Expire(ctx, r.ID); err != nil {
		return err
	}
	slog.Info("reservation expired", "reservation_id", r.ID, "user_id", r.UserID, "server_id", r.ServerID, "username", usr.Username)
	msg := fmt.Sprintf("Reservation expired: user %s SSH access to %s has been revoked", usr.Username, srv.Name)
	if err := s.slack.Notify(ctx, msg); err != nil {
		slog.Warn("slack notify failed", "reservation_id", r.ID, "error", err)
	}
	return nil
}
