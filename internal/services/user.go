package services

import (
	"context"
	"database/sql"

	"github.com/rusik69/serverscheduler/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserServiceDB implements UserService
type UserServiceDB struct {
	db *sql.DB
}

// NewUserService creates a UserService
func NewUserService(db *sql.DB) UserService {
	return &UserServiceDB{db: db}
}

func (s *UserServiceDB) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, COALESCE(ssh_public_key,''), created_at FROM users WHERE username = ?`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.SSHPublicKey, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserServiceDB) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, COALESCE(ssh_public_key,''), created_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.SSHPublicKey, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserServiceDB) Create(ctx context.Context, username, password string, role string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)`,
		username, string(hash), role,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, id)
}

// CreateWithSSHKey creates a user with role=user, password, and SSH public key
func (s *UserServiceDB) CreateWithSSHKey(ctx context.Context, username, password, sshPublicKey string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (username, password_hash, role, ssh_public_key) VALUES (?, ?, 'user', ?)`,
		username, string(hash), sshPublicKey,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, id)
}

func (s *UserServiceDB) UpdateSSHKey(ctx context.Context, username, sshPublicKey string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET ssh_public_key = ? WHERE username = ?`,
		sshPublicKey, username,
	)
	return err
}

func (s *UserServiceDB) List(ctx context.Context) ([]models.UserPublic, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, username, role, created_at FROM users ORDER BY username`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.UserPublic
	for rows.Next() {
		var u models.UserPublic
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (s *UserServiceDB) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

var ErrUserNotFound = &userError{msg: "user not found"}

type userError struct{ msg string }

func (e *userError) Error() string { return e.msg }
