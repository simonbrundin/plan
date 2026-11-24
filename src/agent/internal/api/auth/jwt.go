package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

// Config holds JWT configuration
type Config struct {
	SecretKey     string        // Secret key for signing
	Issuer        string        // Token issuer
	Audience      string        // Expected audience
	Expiry        time.Duration // Token expiry (default: 24 hours)
	RefreshExpiry time.Duration // Refresh token expiry (default: 7 days)
}

// DefaultConfig returns a default JWT configuration
func DefaultConfig() *Config {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "change-me-in-production-use-strong-secret"
	}

	return &Config{
		SecretKey:     secretKey,
		Issuer:        "pi-agent",
		Audience:      "pi-agent-users",
		Expiry:        24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
}

// Claims represents JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID    string   `json:"user_id,omitempty"`
	Email     string   `json:"email,omitempty"`
	Roles     []string `json:"roles,omitempty"`
	SessionID string   `json:"session_id,omitempty"`
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"` // "Bearer"
}

// Generator generates JWT tokens
type Generator struct {
	config *Config
}

// NewGenerator creates a new JWT token generator
func NewGenerator(config *Config) *Generator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Generator{config: config}
}

// GenerateTokenPair generates an access and refresh token pair
func (g *Generator) GenerateTokenPair(userID, email string, roles []string) (*TokenPair, error) {
	sessionID := uuid.New().String()

	// Generate access token
	accessToken, err := g.generateAccessToken(userID, email, roles, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := g.generateRefreshToken(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(g.config.Expiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// generateAccessToken generates an access token
func (g *Generator) generateAccessToken(userID, email string, roles []string, sessionID string) (string, error) {
	now := time.Now()

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.config.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{g.config.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(g.config.Expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		SessionID: sessionID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.config.SecretKey))
}

// generateRefreshToken generates a refresh token
func (g *Generator) generateRefreshToken(userID, sessionID string) (string, error) {
	now := time.Now()

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    g.config.Issuer,
			Subject:   userID,
			Audience:  jwt.ClaimStrings{g.config.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(g.config.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		SessionID: sessionID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.config.SecretKey))
}

// Validator validates JWT tokens
type Validator struct {
	config *Config
}

// NewValidator creates a new JWT validator
func NewValidator(config *Config) *Validator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Validator{config: config}
}

// ValidateAccessToken validates an access token and returns claims
func (v *Validator) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Validate issuer and audience
	if claims.Issuer != v.config.Issuer {
		return nil, fmt.Errorf("%w: invalid issuer", ErrInvalidClaims)
	}

	// Check audience manually
	hasAudience := false
	for _, aud := range claims.Audience {
		if aud == v.config.Audience {
			hasAudience = true
			break
		}
	}
	if !hasAudience {
		return nil, fmt.Errorf("%w: invalid audience", ErrInvalidClaims)
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (v *Validator) ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	// Refresh tokens only need user_id and session_id
	if claims.Subject == "" {
		return nil, fmt.Errorf("%w: missing subject", ErrInvalidClaims)
	}

	return claims, nil
}

// RefreshTokenPair refreshes a token pair using a refresh token
func (g *Generator) RefreshTokenPair(refreshToken string) (*TokenPair, error) {
	validator := NewValidator(g.config)

	claims, err := validator.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair with same session
	return g.GenerateTokenPair(claims.Subject, claims.Email, claims.Roles)
}

// HasRole checks if claims include a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if claims include any of the specified roles
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if claims include all of the specified roles
func (c *Claims) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}

// UserClaims extracts user-relevant claims
type UserClaims struct {
	UserID    string
	Email     string
	Roles     []string
	SessionID string
}

// ToUserClaims converts Claims to UserClaims
func (c *Claims) ToUserClaims() *UserClaims {
	return &UserClaims{
		UserID:    c.UserID,
		Email:     c.Email,
		Roles:     c.Roles,
		SessionID: c.SessionID,
	}
}

// SafeCompare performs constant-time string comparison
func SafeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
