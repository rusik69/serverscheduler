package services

import (
	"context"
	"time"

	"github.com/rusik69/serverscheduler/internal/models"
)

// UserService handles user operations
type UserService interface {
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	Create(ctx context.Context, username, password string, role string) (*models.User, error)
	CreateWithSSHKey(ctx context.Context, username, password, sshPublicKey string) (*models.User, error)
	UpdateSSHKey(ctx context.Context, username, sshPublicKey string) error
	List(ctx context.Context) ([]models.UserPublic, error)
	Delete(ctx context.Context, id int64) error
}

// ServerService handles server operations
type ServerService interface {
	List(ctx context.Context) ([]models.Server, error)
	Get(ctx context.Context, id int64) (*models.Server, error)
	Create(ctx context.Context, s *models.Server) (*models.Server, error)
	Delete(ctx context.Context, id int64) error
}

// ReservationService handles reservation operations
type ReservationService interface {
	Get(ctx context.Context, id int64) (*models.Reservation, error)
	List(ctx context.Context, userID *int64) ([]models.ReservationWithDetails, error)
	Create(ctx context.Context, userID, serverID int64, start, end time.Time) (*models.Reservation, error)
	Cancel(ctx context.Context, id, userID int64) error
	CancelByAdmin(ctx context.Context, id int64) error
	DeleteByUserID(ctx context.Context, userID int64) error
	GetPendingToActivate(ctx context.Context) ([]models.Reservation, error)
	GetActiveToExpire(ctx context.Context) ([]models.Reservation, error)
	Activate(ctx context.Context, id int64) error
	Expire(ctx context.Context, id int64) error
	GetUsersByServer(ctx context.Context) (map[int64][]string, error)
	GetCurrentByServer(ctx context.Context) (map[int64]*models.ReservationWithDetails, error)
}

// SSHService manages SSH keys on remote servers
type SSHService interface {
	AddKey(ctx context.Context, hostname string, port int, sshUser, privateKey, publicKey string) error
	RemoveKey(ctx context.Context, hostname string, port int, sshUser, privateKey, publicKey string) error
	TestConnection(ctx context.Context, hostname string, port int, sshUser, privateKey string) error
}

// SlackService sends notifications to Slack
type SlackService interface {
	Notify(ctx context.Context, message string) error
}
