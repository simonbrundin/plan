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
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/dev-login", frontendBaseURL()))
		return
	}

	zitadelDomain := osGetenv("ZITADEL_DOMAIN", "")
	clientID := osGetenv("ZITADEL_CLIENT_ID", "")
	worktree := osGetenv("WORKTREE_NAME", "")

	// OAuth callback URL. Must be identical to the one used in the token
	// exchange, otherwise Zitadel rejects the exchange. Configurable via
	// ZITADEL_REDIRECT_URI so dev can point at its own address.
	redirectURI := oauthRedirectURI()

	if worktree != "" {
		// Mark this browser as belonging to the current worktree so the
		// dev gateway can route the callback to the correct upstream.
		c.SetCookie("wt", worktree, 600, "/", "", c.Request.TLS != nil, false)
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
	redirectURI := oauthRedirectURI()

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

	// The session lives in the Go API: the browser gets the session cookie
	// directly from this response, so the token never travels in a URL.
	middleware.SetSessionCookies(c, sessionToken, sub, email)

	c.Redirect(http.StatusTemporaryRedirect, frontendRootURL())
}

// frontendRootURL returns the URL of the app the user should land on after
// a successful login. In worktree mode the app lives under the worktree
// path of the shared dev gateway.
func frontendRootURL() string {
	base := frontendBaseURL()

	if worktree := osGetenv("WORKTREE_NAME", ""); worktree != "" && os.Getenv("FRONTEND_URL") == "" {
		return fmt.Sprintf("%s/%s/", base, worktree)
	}

	return fmt.Sprintf("%s/", base)
}

// oauthRedirectURI returns the redirect_uri that is sent to Zitadel in the
// authorize request. The exact same value must be used in the token
// exchange, so both callers use this helper.
func oauthRedirectURI() string {
	if uri := os.Getenv("ZITADEL_REDIRECT_URI"); uri != "" {
		return uri
	}

	domain := osGetenv("APP_DOMAIN", "plan.simonbrundin.com")
	if worktree := osGetenv("WORKTREE_NAME", ""); worktree != "" {
		// The shared dev gateway owns the callback and routes it to the
		// right worktree using the `wt` cookie.
		return fmt.Sprintf("http://%s/oauth/callback", domain)
	}

	return fmt.Sprintf("https://%s/api/v1/auth/callback", domain)
}

// frontendBaseURL returns the base URL of the Nuxt frontend that the user
// should end up on after a successful login.
func frontendBaseURL() string {
	if base := os.Getenv("FRONTEND_URL"); base != "" {
		return strings.TrimSuffix(base, "/")
	}

	domain := osGetenv("APP_DOMAIN", "plan.simonbrundin.com")
	if worktree := osGetenv("WORKTREE_NAME", ""); worktree != "" {
		return fmt.Sprintf("http://%s", domain)
	}

	return fmt.Sprintf("https://%s", domain)
}

// DevLogin creates a session for a local development user without going
// through Zitadel. It is only available when DEV_AUTH_ENABLED=true, which
// is set in the dev environment only.
func (h *AuthHandler) DevLogin(c *gin.Context) {
	if os.Getenv("DEV_AUTH_ENABLED") != "true" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	sub := osGetenv("DEV_AUTH_SUB", "dev-user")
	email := osGetenv("DEV_AUTH_EMAIL", "dev@plan.local")

	userID, err := middleware.LookupOrCreateUser(sub, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	sessionToken, err := middleware.CreateSession(userID, sub, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	middleware.SetSessionCookies(c, sessionToken, sub, email)

	c.Redirect(http.StatusTemporaryRedirect, frontendRootURL())
}

// Session returns the user behind the current session (cookie or Bearer).
func (h *AuthHandler) Session(c *gin.Context) {
	token := middleware.SessionTokenFromRequest(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	session, err := middleware.ValidateSession(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sub":     session.Sub,
		"email":   session.Email,
		"user_id": session.UserID,
	})
}

// Logout deletes the current session and clears the cookies.
func (h *AuthHandler) Logout(c *gin.Context) {
	if token := middleware.SessionTokenFromRequest(c); token != "" {
		middleware.DeleteSession(token)
	}

	middleware.ClearSessionCookies(c)

	c.JSON(http.StatusOK, gin.H{"success": true})
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
