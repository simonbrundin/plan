package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	dbConnected bool
}

func NewAuthHandler(dbConnected bool) *AuthHandler {
	return &AuthHandler{dbConnected: dbConnected}
}

type OAuthState struct {
	State        string `json:"state"`
	PKCEVerifier string `json:"pkce_verifier"`
	RedirectURL  string `json:"redirect_url"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	zitadelDomain := c.GetHeader("X-Zitadel-Domain")
	clientID := c.GetHeader("X-Zitadel-Client-ID")
	redirectURI := c.Query("redirect_uri")

	if zitadelDomain == "" {
		zitadelDomain = osGetenv("ZITADEL_DOMAIN", "")
	}
	if clientID == "" {
		clientID = osGetenv("ZITADEL_CLIENT_ID", "")
	}

	if zitadelDomain == "" || clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Zitadel configuration missing"})
		return
	}

	// Generate PKCE verifier
	verifier := generateRandomString(64)

	// Generate state
	state := generateRandomString(32)

	// Calculate challenge
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// Store state in cookie or return it
	stateData := OAuthState{
		State:        state,
		PKCEVerifier: verifier,
		RedirectURL:  redirectURI,
	}
	stateJSON, _ := json.Marshal(stateData)

	c.SetCookie("oauth_state", string(stateJSON), 600, "/", "", c.Request.TLS != nil, true)

	// Build authorization URL
	authURL := "https://" + zitadelDomain + "/oauth/v2/authorize?" +
		"response_type=code" +
		"&client_id=" + url.QueryEscape(clientID) +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&scope=openid%20email%20profile" +
		"&state=" + url.QueryEscape(state) +
		"&code_challenge=" + url.QueryEscape(challenge) +
		"&code_challenge_method=S256"

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
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

	// Get stored state
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing OAuth state"})
		return
	}

	var storedState OAuthState
	if err := json.Unmarshal([]byte(stateCookie), &storedState); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
		return
	}

	if storedState.State != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State mismatch"})
		return
	}

	// Clear cookie
	c.SetCookie("oauth_state", "", -1, "/", "", c.Request.TLS != nil, true)

	zitadelDomain := osGetenv("ZITADEL_DOMAIN", "")
	clientID := osGetenv("ZITADEL_CLIENT_ID", "")
	clientSecret := osGetenv("ZITADEL_CLIENT_SECRET", "")
	redirectURI := storedState.RedirectURL

	// Exchange code for token
	tokenURL := "https://" + zitadelDomain + "/oauth/v2/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("code_verifier", storedState.PKCEVerifier)

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
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

	// Get user info
	accessToken := tokenResponse["access_token"].(string)
	userInfoURL := "https://" + zitadelDomain + "/oidc/v1/userinfo"

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

	// Return token and user info
	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"token_type":   tokenResponse["token_type"],
		"expires_in":   tokenResponse["expires_in"],
		"user": gin.H{
			"sub":                userInfo["sub"],
			"email":              userInfo["email"],
			"name":               userInfo["name"],
			"preferred_username": userInfo["preferred_username"],
		},
	})
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
