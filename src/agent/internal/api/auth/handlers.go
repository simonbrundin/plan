package auth

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	generator *Generator
	validator *Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(config *Config) *AuthHandler {
	return &AuthHandler{
		generator: NewGenerator(config),
		validator: NewValidator(config),
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	TokenPair
	User *UserInfo `json:"user,omitempty"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID    string   `json:"id"`
	Email string   `json:"email"`
	Roles []string `json:"roles,omitempty"`
}

// TokenResponse is the response for token endpoints
type TokenResponse struct {
	TokenPair
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest represents a logout request (for token invalidation tracking)
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// TokenBlacklist stores invalidated tokens (in production, use Redis)
type TokenBlacklist struct {
	tokens map[string]time.Time
}

// NewTokenBlacklist creates a new token blacklist
func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}
}

// Add adds a token to the blacklist until its expiry
func (b *TokenBlacklist) Add(tokenID string, expiry time.Time) {
	b.tokens[tokenID] = expiry
}

// IsBlacklisted checks if a token is blacklisted
func (b *TokenBlacklist) IsBlacklisted(tokenID string) bool {
	expiry, exists := b.tokens[tokenID]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(b.tokens, tokenID)
		return false
	}
	return true
}

// Cleanup removes expired entries
func (b *TokenBlacklist) Cleanup() {
	now := time.Now()
	for id, expiry := range b.tokens {
		if now.After(expiry) {
			delete(b.tokens, id)
		}
	}
}

// Global token blacklist
var globalBlacklist = NewTokenBlacklist()

// Login authenticates a user and returns tokens
// @Summary User login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "BAD_REQUEST",
		})
		return
	}

	// In production, validate against a user database
	// For now, we use a simple mock authentication
	user, err := h.authenticateUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Generate tokens
	tokenPair, err := h.generator.GenerateTokenPair(user.ID, user.Email, user.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate tokens",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		TokenPair: *tokenPair,
		User: &UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Roles: user.Roles,
		},
	})
}

// authenticateUser validates credentials and returns the user
// In production, this would check against a database
func (h *AuthHandler) authenticateUser(email, password string) (*UserInfo, error) {
	// Mock user database - in production, check against real user store
	validUsers := map[string]struct {
		Password string
		User     *UserInfo
	}{
		"admin@example.com": {
			Password: os.Getenv("ADMIN_PASSWORD"),
			User: &UserInfo{
				ID:    "user-001",
				Email: "admin@example.com",
				Roles: []string{"admin", "user"},
			},
		},
		"user@example.com": {
			Password: "password123",
			User: &UserInfo{
				ID:    "user-002",
				Email: "user@example.com",
				Roles: []string{"user"},
			},
		},
	}

	// Use constant-time comparison
	user, exists := validUsers[email]
	if !exists {
		return nil, ErrInvalidClaims
	}

	if !SafeCompare(user.Password, password) {
		return nil, ErrInvalidClaims
	}

	return user.User, nil
}

// Refresh refreshes access tokens using a refresh token
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
			"code":  "BAD_REQUEST",
		})
		return
	}

	tokenPair, err := h.generator.RefreshTokenPair(req.RefreshToken)
	if err != nil {
		status := http.StatusUnauthorized
		message := "invalid refresh token"
		code := "UNAUTHORIZED"

		if err == ErrExpiredToken {
			message = "refresh token has expired"
			code = "TOKEN_EXPIRED"
		}

		c.JSON(status, gin.H{
			"error": message,
			"code":  code,
		})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{TokenPair: *tokenPair})
}

// Logout invalidates tokens
// @Summary User logout
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LogoutRequest false "Logout request"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	claims, ok := GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	// Add access token to blacklist if we have the jti
	if claims.ID != "" {
		globalBlacklist.Add(claims.ID, claims.ExpiresAt.Time)
	}

	// Blacklist refresh token if provided
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		refreshClaims, err := h.validator.ValidateRefreshToken(req.RefreshToken)
		if err == nil && refreshClaims.ID != "" {
			globalBlacklist.Add(refreshClaims.ID, refreshClaims.ExpiresAt.Time)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "logged out",
	})
}

// Me returns the current user's information
// @Summary Get current user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfo
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := GetUserClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	c.JSON(http.StatusOK, UserInfo{
		ID:    claims.UserID,
		Email: claims.Email,
		Roles: claims.Roles,
	})
}

// IsTokenBlacklisted checks if the current token is blacklisted
func IsTokenBlacklisted(tokenID string) bool {
	return globalBlacklist.IsBlacklisted(tokenID)
}
