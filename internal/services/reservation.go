package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/rusik69/serverscheduler/internal/models"
)

// ReservationServiceDB implements ReservationService
type ReservationServiceDB struct {
	db *sql.DB
}

// NewReservationService creates a ReservationService
func NewReservationService(db *sql.DB) ReservationService {
	return &ReservationServiceDB{db: db}
}

func (s *ReservationServiceDB) Get(ctx context.Context, id int64) (*models.Reservation, error) {
	var r models.Reservation
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, server_id, start_time, end_time, status, created_at FROM reservations WHERE id = ?`,
		id,
	).Scan(&r.ID, &r.UserID, &r.ServerID, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *ReservationServiceDB) List(ctx context.Context, userID *int64) ([]models.ReservationWithDetails, error) {
	var rows *sql.Rows
	var err error
	if userID != nil {
		rows, err = s.db.QueryContext(ctx,
			`SELECT r.id, r.user_id, r.server_id, r.start_time, r.end_time, r.status, r.created_at, s.name, u.username
			 FROM reservations r
			 JOIN servers s ON r.server_id = s.id
			 JOIN users u ON r.user_id = u.id
			 WHERE r.user_id = ?
			 ORDER BY r.start_time DESC`,
			*userID,
		)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT r.id, r.user_id, r.server_id, r.start_time, r.end_time, r.status, r.created_at, s.name, u.username
			 FROM reservations r
			 JOIN servers s ON r.server_id = s.id
			 JOIN users u ON r.user_id = u.id
			 ORDER BY r.start_time DESC`,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.ReservationWithDetails
	for rows.Next() {
		var r models.ReservationWithDetails
		if err := rows.Scan(&r.ID, &r.UserID, &r.ServerID, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt, &r.ServerName, &r.Username); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *ReservationServiceDB) Create(ctx context.Context, userID, serverID int64, start, end time.Time) (*models.Reservation, error) {
	// Validate no overlap for same server (one user at a time per server)
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM reservations WHERE server_id = ? AND status IN ('pending','active')
		 AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))`,
		serverID, end, start, end, start,
	).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrOverlap
	}

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO reservations (user_id, server_id, start_time, end_time, status) VALUES (?, ?, ?, ?, 'pending')`,
		userID, serverID, start, end,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	var r models.Reservation
	err = s.db.QueryRowContext(ctx,
		`SELECT id, user_id, server_id, start_time, end_time, status, created_at FROM reservations WHERE id = ?`,
		id,
	).Scan(&r.ID, &r.UserID, &r.ServerID, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *ReservationServiceDB) Cancel(ctx context.Context, id, userID int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE reservations SET status = 'cancelled' WHERE id = ? AND user_id = ? AND status IN ('pending','active')`,
		id, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *ReservationServiceDB) CancelByAdmin(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE reservations SET status = 'cancelled' WHERE id = ? AND status IN ('pending','active')`,
		id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *ReservationServiceDB) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM reservations WHERE user_id = ?`, userID)
	return err
}

func (s *ReservationServiceDB) GetPendingToActivate(ctx context.Context) ([]models.Reservation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, server_id, start_time, end_time, status, created_at FROM reservations
		 WHERE status = 'pending' AND start_time <= datetime('now') ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReservations(rows)
}

func (s *ReservationServiceDB) GetActiveToExpire(ctx context.Context) ([]models.Reservation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, server_id, start_time, end_time, status, created_at FROM reservations
		 WHERE status = 'active' AND end_time <= datetime('now') ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReservations(rows)
}

func (s *ReservationServiceDB) Activate(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE reservations SET status = 'active' WHERE id = ?`, id)
	return err
}

func (s *ReservationServiceDB) Expire(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE reservations SET status = 'expired' WHERE id = ?`, id)
	return err
}

func scanReservations(rows *sql.Rows) ([]models.Reservation, error) {
	var list []models.Reservation
	for rows.Next() {
		var r models.Reservation
		if err := rows.Scan(&r.ID, &r.UserID, &r.ServerID, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// GetCurrentByServer returns the most relevant reservation per server:
// active first, then the nearest upcoming pending.
func (s *ReservationServiceDB) GetCurrentByServer(ctx context.Context) (map[int64]*models.ReservationWithDetails, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.user_id, r.server_id, r.start_time, r.end_time, r.status, r.created_at, u.username
		 FROM reservations r
		 JOIN users u ON r.user_id = u.id
		 WHERE r.status IN ('active','pending') AND r.end_time > datetime('now')
		 ORDER BY
		   CASE r.status WHEN 'active' THEN 0 ELSE 1 END,
		   r.start_time ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64]*models.ReservationWithDetails)
	for rows.Next() {
		var r models.ReservationWithDetails
		if err := rows.Scan(&r.ID, &r.UserID, &r.ServerID, &r.StartTime, &r.EndTime, &r.Status, &r.CreatedAt, &r.Username); err != nil {
			return nil, err
		}
		if _, exists := m[r.ServerID]; !exists {
			r2 := r
			m[r.ServerID] = &r2
		}
	}
	return m, rows.Err()
}

func (s *ReservationServiceDB) GetUsersByServer(ctx context.Context) (map[int64][]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT r.server_id, u.username
		 FROM reservations r
		 JOIN users u ON r.user_id = u.id
		 WHERE r.status IN ('pending', 'active')
		 ORDER BY r.server_id, u.username`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64][]string)
	for rows.Next() {
		var serverID int64
		var username string
		if err := rows.Scan(&serverID, &username); err != nil {
			return nil, err
		}
		m[serverID] = append(m[serverID], username)
	}
	return m, rows.Err()
}

// ErrOverlap and ErrNotFound for reservation errors
var ErrOverlap = &reservationError{msg: "reservation overlaps with existing one"}
var ErrNotFound = &reservationError{msg: "reservation not found"}

type reservationError struct{ msg string }

func (e *reservationError) Error() string { return e.msg }
