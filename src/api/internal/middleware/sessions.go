package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	UserID    int64
	Sub       string
	Email     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore manages active sessions
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	secret   []byte
}

var globalSessionStore = &SessionStore{
	sessions: make(map[string]*Session),
	secret:   []byte(os.Getenv("SESSION_SECRET")),
}

// InitSessionStore initializes the session store with a secret
func InitSessionStore(secret string) {
	globalSessionStore.mu.Lock()
	defer globalSessionStore.mu.Unlock()
	if secret != "" {
		globalSessionStore.secret = []byte(secret)
	} else {
		// Generate a random secret if not provided
		// In production, this should be set via environment variable
		globalSessionStore.secret = []byte("dev-secret-change-in-production")
	}
}

// CreateSession creates a new session and returns the session token
func CreateSession(userID int64, sub, email string) (string, error) {
	if len(globalSessionStore.secret) == 0 {
		InitSessionStore("")
	}

	session := &Session{
		UserID:    userID,
		Sub:       sub,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Sessions last 24 hours
	}

	// Generate session token
	token := generateSessionToken(session)

	globalSessionStore.mu.Lock()
	globalSessionStore.sessions[token] = session
	globalSessionStore.mu.Unlock()

	return token, nil
}

// ValidateSession validates a session token and returns the session
func ValidateSession(token string) (*Session, error) {
	globalSessionStore.mu.RLock()
	session, exists := globalSessionStore.sessions[token]
	globalSessionStore.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		globalSessionStore.mu.Lock()
		delete(globalSessionStore.sessions, token)
		globalSessionStore.mu.Unlock()
		return nil, fmt.Errorf("session expired")
	}

	// Verify token integrity
	expectedToken := generateSessionToken(session)
	if !hmac.Equal([]byte(token), []byte(expectedToken)) {
		return nil, fmt.Errorf("invalid session token")
	}

	return session, nil
}

// DeleteSession removes a session
func DeleteSession(token string) {
	globalSessionStore.mu.Lock()
	delete(globalSessionStore.sessions, token)
	globalSessionStore.mu.Unlock()
}

// CleanupExpiredSessions removes expired sessions
func CleanupExpiredSessions() {
	globalSessionStore.mu.Lock()
	defer globalSessionStore.mu.Unlock()

	now := time.Now()
	for token, session := range globalSessionStore.sessions {
		if now.After(session.ExpiresAt) {
			delete(globalSessionStore.sessions, token)
		}
	}
}

// generateSessionToken creates a signed session token
func generateSessionToken(session *Session) string {
	data := map[string]interface{}{
		"user_id":    session.UserID,
		"sub":        session.Sub,
		"email":      session.Email,
		"created_at": session.CreatedAt.Unix(),
		"expires_at": session.ExpiresAt.Unix(),
	}

	jsonData, _ := json.Marshal(data)
	encoded := base64.RawURLEncoding.EncodeToString(jsonData)

	// Create HMAC signature
	h := hmac.New(sha256.New, globalSessionStore.secret)
	h.Write(jsonData)
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s.%s", encoded, signature)
}

// StartSessionCleanup starts a background goroutine to clean up expired sessions
func StartSessionCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			CleanupExpiredSessions()
		}
	}()
}
