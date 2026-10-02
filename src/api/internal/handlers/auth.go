package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"plan-api/internal/middleware"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	dbConnected bool
}

func NewAuthHandler(dbConnected bool) *AuthHandler {
	return &AuthHandler{dbConnected: dbConnected}
}

type LoginResponse struct {
	AuthURL string `json:"auth_url"`
}

type TokenResponse struct {
	AccessToken string   `json:"access_token"`
	User        UserInfo `json:"user"`
}

type UserInfo struct {
	Sub    string `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	UserID int64  `json:"user_id"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	// Dev mode: redirect to Nuxt dev-login endpoint
	if os.Getenv("AUTH_DISABLED") == "true" {
		appDomain := osGetenv("APP_DOMAIN", "localhost:3000")
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("http://%s/auth/dev-login", appDomain))
		return
	}

	zitadelDomain := osGetenv("ZITADEL_DOMAIN", "")
	clientID := osGetenv("ZITADEL_CLIENT_ID", "")
	worktree := osGetenv("WORKTREE_NAME", "")

	// OAuth callback URL. In dev (worktree mode) it always points to the
	// shared dev gateway so multiple worktrees can share one Zitadel
	// application. In prod it points to the worktree's host directly.
	var redirectURI string
	if worktree != "" {
		appDomain := osGetenv("APP_DOMAIN", "localhost:3000")
		redirectURI = fmt.Sprintf("http://%s/oauth/callback", appDomain)
		// Mark this browser as belonging to the current worktree so the
		// dev gateway can route the callback to the correct upstream.
		c.SetCookie("wt", worktree, 600, "/", "", c.Request.TLS != nil, false)
	} else {
		redirectURI = fmt.Sprintf("https://%s/api/v1/auth/callback", osGetenv("APP_DOMAIN", "plan.simonbrundin.com"))
	}

	if zitadelDomain == "" || clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zitadel configuration missing"})
		return
	}

	// Generate PKCE
	verifier := generateRandomString(64)
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	state := generateRandomString(32)

	// Store state in cookie
	stateData := map[string]string{
		"state":    state,
		"verifier": verifier,
		"worktree": worktree,
	}
	stateJSON, _ := json.Marshal(stateData)
	c.SetCookie("oauth_state", string(stateJSON), 600, "/", "", c.Request.TLS != nil, true)

	// Build auth URL
	authURL := fmt.Sprintf(
		"https://%s/oauth/v2/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=openid%%20email%%20profile&state=%s&code_challenge=%s&code_challenge_method=S256",
		zitadelDomain,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
		url.QueryEscape(challenge),
	)

	// Redirect to Zitadel
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       errorParam,
			"description": c.Query("error_description"),
		})
		return
	}

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code or state"})
		return
	}

	// Validate state
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing OAuth state"})
		return
	}

	var storedState map[string]string
	if err := json.Unmarshal([]byte(stateCookie), &storedState); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
		return
	}

	if storedState["state"] != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State mismatch"})
		return
	}

	c.SetCookie("oauth_state", "", -1, "/", "", c.Request.TLS != nil, true)

	zitadelDomain := osGetenv("ZITADEL_DOMAIN", "")
	clientID := osGetenv("ZITADEL_CLIENT_ID", "")
	clientSecret := osGetenv("ZITADEL_CLIENT_SECRET", "")
	redirectURI := fmt.Sprintf("https://%s/api/v1/auth/callback", osGetenv("APP_DOMAIN", "plan.simonbrundin.com"))

	// Exchange code for token
	tokenURL := fmt.Sprintf("https://%s/oauth/v2/token", zitadelDomain)

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", storedState["verifier"])

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Token exchange failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token response"})
		return
	}

	if tokenResponse["error"] != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       tokenResponse["error"],
			"description": tokenResponse["error_description"],
		})
		return
	}

	accessToken := tokenResponse["access_token"].(string)

	// Get user info
	userInfoURL := fmt.Sprintf("https://%s/oidc/v1/userinfo", zitadelDomain)
	userReq, _ := http.NewRequest("GET", userInfoURL, nil)
	userReq.Header.Set("Authorization", "Bearer "+accessToken)

	userResp, err := client.Do(userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User info request failed"})
		return
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)
	var userInfo map[string]interface{}
	if err := json.Unmarshal(userBody, &userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user info"})
		return
	}

	sub, _ := userInfo["sub"].(string)
	email, _ := userInfo["email"].(string)

	if sub == "" || email == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing user data"})
		return
	}

	// Create or find user in database
	userID, err := middleware.LookupOrCreateUser(sub, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user session"})
		return
	}

	// Create our own session token (not the Zitadel token)
	sessionToken, err := middleware.CreateSession(userID, sub, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Redirect to Nuxt with our session token.
	// In worktree/dev mode the callback URL is scoped to the worktree path
	// so the dev gateway routes it to the correct Nuxt instance.
	frontendCallbackURL := buildFrontendCallbackURL(sessionToken, sub, email)

	c.Redirect(http.StatusTemporaryRedirect, frontendCallbackURL)
}

// buildFrontendCallbackURL returns the URL the Go API redirects to after a
// successful Zitadel callback. In worktree/dev mode the URL is scoped to
// the worktree path so the shared dev gateway (Caddy) can route the
// request to the correct Nuxt instance. In prod the URL is the plain
// per-domain Nuxt callback.
func buildFrontendCallbackURL(sessionToken, sub, email string) string {
	appDomain := osGetenv("APP_DOMAIN", "plan.simonbrundin.com")
	worktree := osGetenv("WORKTREE_NAME", "")

	var scheme string
	if worktree != "" {
		scheme = "http"
	} else {
		scheme = "https"
	}

	var callbackPath string
	if worktree != "" {
		callbackPath = fmt.Sprintf("/%s/api/auth/callback", worktree)
	} else {
		callbackPath = "/api/auth/callback"
	}

	return fmt.Sprintf(
		"%s://%s%s?session=%s&sub=%s&email=%s",
		scheme,
		appDomain,
		callbackPath,
		url.QueryEscape(sessionToken),
		url.QueryEscape(sub),
		url.QueryEscape(email),
	)
}

// RefreshSession handles session refresh requests
func (h *AuthHandler) RefreshSession(c *gin.Context) {
	sessionToken := c.GetHeader("X-Session-Token")
	if sessionToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session token required"})
		return
	}

	session, err := middleware.ValidateSession(sessionToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	// Refresh the session (extend expiry)
	middleware.DeleteSession(sessionToken)
	newToken, err := middleware.CreateSession(session.UserID, session.Sub, session.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": newToken})
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}

func osGetenv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
