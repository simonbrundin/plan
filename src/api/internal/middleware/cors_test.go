package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Browsaren anropar API:t direkt - i dev ligger appen på :3100 och API:t på
// :4100, så API:t måste svara med CORS-headrar för appens origin.
//
// Viktigt: Access-Control-Allow-Origin får inte vara "*" när anropet bär
// sessioncookien - webbläsaren blockerar då credentialed requests. Origin
// ska därför echas tillbaka.
func TestCORSMiddlewareEchoarTillatenOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3100")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/api/v1/goals", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	t.Run("tillåten origin får CORS-headrar", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/goals", nil)
		req.Header.Set("Origin", "http://localhost:3100")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3100" {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3100")
		}
		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
		}
		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Errorf("Vary = %q, want %q", got, "Origin")
		}
	})

	t.Run("otillåten origin får inga CORS-headrar", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/goals", nil)
		req.Header.Set("Origin", "http://evil.example")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
			t.Errorf("Access-Control-Allow-Credentials = %q, want empty", got)
		}
	})

	t.Run("utan Origin-headrar svarar API:t utan CORS", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/goals", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
		}
	})

	t.Run("preflight svarar 204 med headrar", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/goals", nil)
		req.Header.Set("Origin", "http://localhost:3100")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3100" {
			t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3100")
		}
	})
}

// Flera origin kan listas - appen kan ligga på flera adresser.
func TestCORSMiddlewareGodtarFleraOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3100, https://plan.simonbrundin.com")

	if !originAllowed("http://localhost:3100") {
		t.Error("http://localhost:3100 ska vara tillåten")
	}
	if !originAllowed("https://plan.simonbrundin.com") {
		t.Error("https://plan.simonbrundin.com ska vara tillåten")
	}
	if originAllowed("https://plan.simonbrundin.com:8443") {
		t.Error("en annan port ska inte vara tillåten")
	}
	if originAllowed("http://localhost:3101") {
		t.Error("en annan port ska inte vara tillåten")
	}
}

// "*" tillåter alla origin - men då echas origin tillbaka, aldrig wildcard,
// eftersom anropet bär sessioncookien.
func TestCORSMiddlewareWildcardEchasOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "*")

	if !originAllowed("http://localhost:3100") {
		t.Error("wildcard ska tillåta alla origin")
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/api/v1/goals", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/goals", nil)
	req.Header.Set("Origin", "http://localhost:3100")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3100" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3100")
	}
}
