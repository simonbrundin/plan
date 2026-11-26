package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Ett session-token måste kunna avkodas tillbaka till exakt samma data.
// Tidigare skrev signSession Unix-timestamps ("created_at": 1234) medan
// Session-structen saknade JSON-taggar, vilket gjorde att ExpiresAt blevi
// nolltiden - varje session uppfattades då som "expired" och API:t svarade
// 401 trots korrekt cookie.
func TestSessionTokenRoundTrip(t *testing.T) {
	session := &Session{
		UserID:    42,
		Sub:       "sub-1",
		Email:     "user@example.com",
		CreatedAt: time.Now().Truncate(time.Second),
		ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Second),
	}

	parsed, err := parseSession(signSession(session))
	if err != nil {
		t.Fatalf("parseSession failed: %v", err)
	}

	if parsed.UserID != session.UserID {
		t.Errorf("UserID = %d, want %d", parsed.UserID, session.UserID)
	}
	if parsed.Sub != session.Sub {
		t.Errorf("Sub = %q, want %q", parsed.Sub, session.Sub)
	}
	if parsed.Email != session.Email {
		t.Errorf("Email = %q, want %q", parsed.Email, session.Email)
	}
	if !parsed.ExpiresAt.After(time.Now()) {
		t.Errorf("ExpiresAt = %v, want a future timestamp", parsed.ExpiresAt)
	}
}

func TestParseSessionRejectsTamperedToken(t *testing.T) {
	session := &Session{
		UserID:    1,
		Sub:       "sub-1",
		Email:     "user@example.com",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	token := signSession(session)
	if _, err := parseSession(token + "x"); err == nil {
		t.Fatal("expected error for tampered token")
	}
	if _, err := parseSession("not-a-token"); err == nil {
		t.Fatal("expected error for malformed token")
	}
}

// Authen ligger i Go API:t - webbläsaren får sessionen via cookie. Både
// cookie och Bearer-header ska därför kunna ge samma token.
func TestSessionTokenFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		cookie string
		header string
		want   string
	}{
		{name: "session cookie", cookie: "plan_session=cookie-token", want: "cookie-token"},
		{name: "bearer header", header: "Bearer header-token", want: "header-token"},
		{name: "cookie takes precedence", cookie: "plan_session=cookie-token", header: "Bearer header-token", want: "cookie-token"},
		{name: "no credentials", want: ""},
		{name: "malformed header", header: "Basic abc", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookie != "" {
				c.Request.Header.Set("Cookie", tt.cookie)
			}
			if tt.header != "" {
				c.Request.Header.Set("Authorization", tt.header)
			}

			if got := SessionTokenFromRequest(c); got != tt.want {
				t.Errorf("SessionTokenFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Sessionen sätts som httpOnly-cookie, och hint-cookien är läsbar för
// klienten (men inte auktoritativ - servern validerar sessionen själv).
func TestSetAndClearSessionCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	SetSessionCookies(c, "token-123", "sub-1", "user@example.com")

	cookies := rec.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("got %d cookies, want 2", len(cookies))
	}

	var sessionCookie, userCookie *http.Cookie
	for _, cookie := range cookies {
		switch cookie.Name {
		case SessionCookieName:
			sessionCookie = cookie
		case UserCookieName:
			userCookie = cookie
		}
	}

	if sessionCookie == nil || sessionCookie.Value != "token-123" {
		t.Fatalf("session cookie = %v, want token-123", sessionCookie)
	}
	if !sessionCookie.HttpOnly {
		t.Error("session cookie must be httpOnly")
	}

	if userCookie == nil {
		t.Fatal("user hint cookie is missing")
	}
	if userCookie.HttpOnly {
		t.Error("user hint cookie must be readable by the client")
	}

	// Utloggning ska rensa båda cookierna.
	rec = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ClearSessionCookies(c)

	for _, cookie := range rec.Result().Cookies() {
		if cookie.Value != "" || cookie.MaxAge >= 0 {
			t.Errorf("cookie %s was not cleared: value=%q maxAge=%d", cookie.Name, cookie.Value, cookie.MaxAge)
		}
	}
}
