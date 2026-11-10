package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"plan-api/internal/database"

	"github.com/gin-gonic/gin"
)

// Session represents a user session
type Session struct {
	UserID    int64     `json:"user_id"`
	Sub       string    `json:"sub"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// dbSessionStore persists sessions in PostgreSQL
type dbSessionStore struct {
	secret []byte
	mu     sync.RWMutex
}

var globalSessionStore *dbSessionStore

func init() {
	// Use NUXT_SESSION_PASSWORD so Nuxt and Go API use the same session secret
	// (nuxt-auth-utils uses NUXT_SESSION_PASSWORD for cookie signing)
	globalSessionStore = &dbSessionStore{
		secret: []byte(os.Getenv("NUXT_SESSION_PASSWORD")),
	}
	if len(globalSessionStore.secret) == 0 {
		// Fallback to SESSION_SECRET for backwards compatibility
		globalSessionStore.secret = []byte(os.Getenv("SESSION_SECRET"))
	}
	if len(globalSessionStore.secret) == 0 {
		globalSessionStore.secret = []byte("dev-secret-change-in-production")
	}
}

// InitSessionStore initializes the session store (no-op for DB store)
func InitSessionStore(secret string) {
	globalSessionStore.mu.Lock()
	defer globalSessionStore.mu.Unlock()
	if secret != "" {
		globalSessionStore.secret = []byte(secret)
	}
}

// GetSecret returns the current session secret (for testing/debugging)
func GetSecret() string {
	return string(globalSessionStore.secret)
}

// ensureSessionsTable creates the sessions table if it doesn't exist
func ensureSessionsTable(ctx context.Context) error {
	_, err := database.GetPool().Exec(ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT PRIMARY KEY,
			user_id   BIGINT NOT NULL,
			sub       TEXT NOT NULL,
			email     TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL
		)
	`)
	return err
}

// CreateSession creates a new session and returns the session token
func CreateSession(userID int64, sub, email string) (string, error) {
	if database.GetPool() == nil {
		return "", fmt.Errorf("database not connected")
	}

	session := &Session{
		UserID:    userID,
		Sub:       sub,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Generate session token
	token := signSession(session)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ensure table exists
	if err := ensureSessionsTable(ctx); err != nil {
		return "", fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Store session in database
	_, err := database.GetPool().Exec(ctx,
		`INSERT INTO sessions (token, user_id, sub, email, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (token) DO UPDATE SET expires_at = $6`,
		token, session.UserID, session.Sub, session.Email, session.CreatedAt, session.ExpiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return token, nil
}

// parseSession decodes a session token and verifies its signature.
// It does not touch the database, which makes it testable on its own.
func parseSession(token string) (*Session, error) {
	parts := splitToken(token)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	dataJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid token encoding")
	}

	var session Session
	if err := json.Unmarshal(dataJSON, &session); err != nil {
		return nil, fmt.Errorf("invalid session data")
	}

	// Verify HMAC
	h := hmac.New(sha256.New, globalSessionStore.secret)
	h.Write(dataJSON)
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid session token")
	}

	return &session, nil
}

const (
	// SessionCookieName is the httpOnly cookie holding the session token.
	SessionCookieName = "plan_session"
	// UserCookieName is a readable hint cookie for the client (sub + email).
	// It is NOT authoritative - the server validates the session itself.
	UserCookieName = "plan_user"
)

// SetSessionCookies stores the session token in an httpOnly cookie and a
// readable hint cookie for the client. The auth flow lives in the Go API,
// so the browser gets its cookies straight from here.
func SetSessionCookies(c *gin.Context, token, sub, email string) {
	secure := c.Request.TLS != nil || os.Getenv("COOKIE_SECURE") == "true"
	maxAge := 60 * 60 * 24 * 7 // 7 days

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})

	userData, _ := json.Marshal(map[string]string{"sub": sub, "email": email})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     UserCookieName,
		Value:    base64.StdEncoding.EncodeToString(userData),
		Path:     "/",
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
}

// ClearSessionCookies removes the session cookies from the browser.
func ClearSessionCookies(c *gin.Context) {
	secure := c.Request.TLS != nil || os.Getenv("COOKIE_SECURE") == "true"

	for _, name := range []string{SessionCookieName, UserCookieName} {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: name == SessionCookieName,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}
}

// SessionTokenFromRequest returns the session token from the session cookie
// or from an Authorization: Bearer header.
func SessionTokenFromRequest(c *gin.Context) string {
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie != "" {
		return cookie
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// ValidateSession validates a session token and returns the session
func ValidateSession(token string) (*Session, error) {
	if database.GetPool() == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// Verify token signature and payload
	session, err := parseSession(token)
	if err != nil {
		return nil, err
	}

	// Check expiry
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.GetPool().Exec(ctx, "DELETE FROM sessions WHERE token = $1", token)
		return nil, fmt.Errorf("session expired")
	}

	// Verify session exists in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var found bool
	err = database.GetPool().QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM sessions WHERE token = $1 AND expires_at > NOW())",
		token,
	).Scan(&found)
	if err != nil || !found {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// DeleteSession removes a session
func DeleteSession(token string) {
	if database.GetPool() == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	database.GetPool().Exec(ctx, "DELETE FROM sessions WHERE token = $1", token)
}

// CleanupExpiredSessions removes expired sessions from database
func CleanupExpiredSessions() {
	if database.GetPool() == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database.GetPool().Exec(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")
}

// signSession creates a signed session token (HMAC-based, no DB needed for the token itself)
func signSession(session *Session) string {
	data := map[string]interface{}{
		"user_id":    session.UserID,
		"sub":        session.Sub,
		"email":      session.Email,
		"created_at": session.CreatedAt.UTC().Format(time.RFC3339),
		"expires_at": session.ExpiresAt.UTC().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(data)
	encoded := base64.RawURLEncoding.EncodeToString(jsonData)

	h := hmac.New(sha256.New, globalSessionStore.secret)
	h.Write(jsonData)
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s.%s", encoded, signature)
}

// splitToken splits a token into data and signature parts
func splitToken(token string) []string {
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			return []string{token[:i], token[i+1:]}
		}
	}
	return nil
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
