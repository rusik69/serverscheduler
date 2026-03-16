package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Session represents an authenticated session
type Session struct {
	Username  string
	ExpiresAt time.Time
}

// SessionStore manages active sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var globalSessionStore = &SessionStore{sessions: make(map[string]*Session)}

// GetSessionStore returns the global session store
func GetSessionStore() *SessionStore {
	return globalSessionStore
}

func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			globalSessionStore.Cleanup()
		}
	}()
}

// GenerateSessionID generates a random session ID
func GenerateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// SetSession creates a new session
func (s *SessionStore) SetSession(sessionID string, username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = &Session{Username: username, ExpiresAt: time.Now().Add(24 * time.Hour)}
}

// GetSession retrieves a session by ID
func (s *SessionStore) GetSession(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, exists := s.sessions[sessionID]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

// DeleteSession removes a session
func (s *SessionStore) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// Cleanup removes expired sessions
func (s *SessionStore) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, id)
		}
	}
}

// AuthMiddleware handles session-based authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicEndpoint(c.Request.URL.Path, c.Request.Method) {
			c.Next()
			return
		}
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		session, exists := globalSessionStore.GetSession(sessionID)
		if !exists {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Set("username", session.Username)
		c.Set("authenticated", true)
		c.Next()
	}
}

func isPublicEndpoint(path, method string) bool {
	_ = method
	public := []string{"/", "/servers", "/login", "/register", "/logout", "/toggle-theme", "/ping", "/health"}
	for _, p := range public {
		if path == p {
			return true
		}
	}
	return false
}

// GetCurrentUser retrieves the current user from context
func GetCurrentUser(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	s, ok := username.(string)
	return s, ok
}
