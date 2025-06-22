package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	user := User{
		ID:       1,
		Username: "testuser",
		Password: "hashedpassword",
		Role:     "user",
	}

	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashedpassword", user.Password)
	assert.Equal(t, "user", user.Role)
}

func TestServerModel(t *testing.T) {
	server := Server{
		ID:        1,
		Name:      "test-server",
		Status:    "available",
		IPAddress: "192.168.1.100",
		Username:  "admin",
		Password:  "secret",
	}

	assert.Equal(t, int64(1), server.ID)
	assert.Equal(t, "test-server", server.Name)
	assert.Equal(t, "available", server.Status)
	assert.Equal(t, "192.168.1.100", server.IPAddress)
	assert.Equal(t, "admin", server.Username)
	assert.Equal(t, "secret", server.Password)
}

func TestReservationModel(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(2 * time.Hour)

	reservation := Reservation{
		ID:         1,
		ServerID:   1,
		UserID:     1,
		ServerName: "test-server",
		Username:   "testuser",
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     "active",
	}

	assert.Equal(t, int64(1), reservation.ID)
	assert.Equal(t, int64(1), reservation.ServerID)
	assert.Equal(t, int64(1), reservation.UserID)
	assert.Equal(t, "test-server", reservation.ServerName)
	assert.Equal(t, "testuser", reservation.Username)
	assert.Equal(t, startTime, reservation.StartTime)
	assert.Equal(t, endTime, reservation.EndTime)
	assert.Equal(t, "active", reservation.Status)
}

func TestUserRoleValidation(t *testing.T) {
	validRoles := []string{"user", "root"}

	for _, role := range validRoles {
		user := User{
			ID:       1,
			Username: "test",
			Password: "pass",
			Role:     role,
		}
		assert.Contains(t, validRoles, user.Role)
	}
}

func TestServerStatusValidation(t *testing.T) {
	validStatuses := []string{"available", "reserved", "maintenance"}

	for _, status := range validStatuses {
		server := Server{
			ID:     1,
			Name:   "test",
			Status: status,
		}
		assert.Contains(t, validStatuses, server.Status)
	}
}

func TestReservationStatusValidation(t *testing.T) {
	validStatuses := []string{"active", "cancelled", "expired"}

	for _, status := range validStatuses {
		reservation := Reservation{
			ID:        1,
			ServerID:  1,
			UserID:    1,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Hour),
			Status:    status,
		}
		assert.Contains(t, validStatuses, reservation.Status)
	}
}
