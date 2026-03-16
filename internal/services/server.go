package services

import (
	"context"
	"database/sql"

	"github.com/rusik69/serverscheduler/internal/models"
)

// ServerServiceDB implements ServerService
type ServerServiceDB struct {
	db *sql.DB
}

// NewServerService creates a ServerService
func NewServerService(db *sql.DB) ServerService {
	return &ServerServiceDB{db: db}
}

func (s *ServerServiceDB) List(ctx context.Context) ([]models.Server, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, hostname, port, ssh_user, description, created_at FROM servers ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Server
	for rows.Next() {
		var sv models.Server
		if err := rows.Scan(&sv.ID, &sv.Name, &sv.Hostname, &sv.Port, &sv.SSHUser, &sv.Description, &sv.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, sv)
	}
	return list, rows.Err()
}

func (s *ServerServiceDB) Get(ctx context.Context, id int64) (*models.Server, error) {
	var sv models.Server
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, hostname, port, ssh_user, ssh_private_key, description, created_at FROM servers WHERE id = ?`,
		id,
	).Scan(&sv.ID, &sv.Name, &sv.Hostname, &sv.Port, &sv.SSHUser, &sv.SSHPrivateKey, &sv.Description, &sv.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sv, nil
}

func (s *ServerServiceDB) Create(ctx context.Context, sv *models.Server) (*models.Server, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO servers (name, hostname, port, ssh_user, ssh_private_key, description) VALUES (?, ?, ?, ?, ?, ?)`,
		sv.Name, sv.Hostname, sv.Port, sv.SSHUser, sv.SSHPrivateKey, sv.Description,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.Get(ctx, id)
}

func (s *ServerServiceDB) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM servers WHERE id = ?`, id)
	return err
}
