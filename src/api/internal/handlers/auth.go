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
	zitadelDomain := osGetenv("ZITADEL_DOMAIN", "")
	clientID := osGetenv("ZITADEL_CLIENT_ID", "")
	redirectURI := fmt.Sprintf("https://%s/api/v1/auth/callback", osGetenv("APP_DOMAIN", "plan.simonbrundin.com"))

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

	// TODO: Create or find user in DB when dbConnected
	_ = h.dbConnected // Suppress unused warning

	// Redirect to frontend callback with token
	frontendCallbackURL := fmt.Sprintf(
		"https://%s/auth/callback?token=%s&sub=%s&email=%s",
		osGetenv("APP_DOMAIN", "plan.simonbrundin.com"),
		url.QueryEscape(accessToken),
		url.QueryEscape(sub),
		url.QueryEscape(email),
	)

	c.Redirect(http.StatusTemporaryRedirect, frontendCallbackURL)
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
