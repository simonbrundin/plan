package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"plan-api/internal/database"
)

// Zitadel introspection response
type introspectionResponse struct {
	Active   bool   `json:"active"`
	Subject  string `json:"sub"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Username string `json:"username"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// JWKS cache
var (
	jwksCache     *JWKS
	jwksCacheMu   sync.RWMutex
	jwksCacheTime time.Time
	jwksCacheTTL  = time.Hour // Refresh JWKS every hour
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// Zitadel JWKS endpoint cache
func getJWKS(zitadelDomain string) (*JWKS, error) {
	jwksCacheMu.RLock()
	if jwksCache != nil && time.Since(jwksCacheTime) < jwksCacheTTL {
		defer jwksCacheMu.RUnlock()
		return jwksCache, nil
	}
	jwksCacheMu.RUnlock()

	// Fetch fresh JWKS
	jwksURL := fmt.Sprintf("https://%s/oauth/v2/keys", zitadelDomain)
	resp, err := http.Get(jwksURL)
	if err != nil {
		// Return cached if available
		jwksCacheMu.RLock()
		defer jwksCacheMu.RUnlock()
		if jwksCache != nil {
			return jwksCache, nil
		}
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		jwksCacheMu.RLock()
		defer jwksCacheMu.RUnlock()
		if jwksCache != nil {
			return jwksCache, nil
		}
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Update cache
	jwksCacheMu.Lock()
	jwksCache = &jwks
	jwksCacheTime = time.Now()
	jwksCacheMu.Unlock()

	return &jwks, nil
}

// GetRSAKey finds the RSA public key for a given key ID
func getRSAKey(zitadelDomain, kid string) (*rsa.PublicKey, error) {
	jwks, err := getJWKS(zitadelDomain)
	if err != nil {
		return nil, err
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid && key.Kty == "RSA" {
			return jwkToRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// Convert JWK to RSA public key
func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

// ZitadelClaims represents the claims in a Zitadel JWT
type ZitadelClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AuthMiddleware validates the Bearer token and extracts user ID
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Development mode: skip authentication
		if os.Getenv("AUTH_DISABLED") == "true" {
			// Default to user 4 (simonbrundin@gmail.com) who has test data
			c.Set("userID", int64(4))
			c.Set("userSub", "378032824856347117")
			c.Set("userEmail", "simonbrundin@gmail.com")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization header required"})
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// For development/testing, accept tokens in format "user_<id>"
		if strings.HasPrefix(tokenString, "user_") {
			var userID int64
			for _, ch := range tokenString[5:] {
				if ch < '0' || ch > '9' {
					log.Printf("Auth: Invalid user token format")
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user token"})
					c.Abort()
					return
				}
				userID = userID*10 + int64(ch-'0')
			}
			log.Printf("Auth: Using dev token for user %d", userID)
			c.Set("userID", userID)
			c.Set("userSub", fmt.Sprintf("user_%d", userID))
			c.Next()
			return
		}

		// Validate Zitadel JWT or opaque token
		zitadelDomain := osGetenv("ZITADEL_DOMAIN", "auth.simonbrundin.com")

		var subject, email string
		var userID int64
		var validationErr error

		// First, try to parse as JWT (for standard JWT tokens)
		parser := jwt.NewParser()
		token, _, err := parser.ParseUnverified(tokenString, &ZitadelClaims{})

		if err == nil && token != nil {
			// Token looks like JWT, try to validate it
			kid, ok := token.Header["kid"].(string)
			if ok {
				publicKey, pubErr := getRSAKey(zitadelDomain, kid)
				if pubErr == nil {
					claims := &ZitadelClaims{}
					validatedToken, valErr := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
						if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
							return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
						}
						return publicKey, nil
					})

					if valErr == nil && validatedToken.Valid {
						log.Printf("Auth: Token validated as JWT, subject=%s, email=%s", claims.Subject, claims.Email)
						subject = claims.Subject
						email = claims.Email
						userID, validationErr = lookupOrCreateUser(subject, email)
						if validationErr == nil {
							c.Set("userID", userID)
							c.Set("userSub", subject)
							c.Set("userEmail", email)
							c.Next()
							return
						}
					}
				}
			}
		}

		// If JWT validation failed, try introspection (for opaque/JWE tokens)
		log.Printf("Auth: JWT parsing failed, trying introspection for token")
		introspectionResult, intErr := introspectToken(zitadelDomain, tokenString)
		if intErr != nil {
			log.Printf("Auth: Introspection failed: %v", intErr)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		if !introspectionResult.Active {
			log.Printf("Auth: Token is not active (expired or revoked)")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token expired or revoked"})
			c.Abort()
			return
		}

		log.Printf("Auth: Token validated via introspection, subject=%s, email=%s",
			introspectionResult.Subject, introspectionResult.Email)

		subject = introspectionResult.Subject
		email = introspectionResult.Email
		if email == "" {
			email = introspectionResult.Username
		}

		userID, validationErr = lookupOrCreateUser(subject, email)
		if validationErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Failed to create user session"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("userSub", subject)
		c.Set("userEmail", email)
		c.Next()
	}
}

// lookupUserBySub looks up a user by their Zitadel subject (sub)
func lookupUserBySub(sub string) (int64, error) {
	pool := database.GetPool()
	if pool == nil {
		log.Printf("Auth: Database pool not available")
		return 0, fmt.Errorf("database not connected")
	}

	var userID int64
	err := pool.QueryRow(context.Background(),
		"SELECT id FROM users WHERE sub = $1", sub).Scan(&userID)
	if err != nil {
		log.Printf("Auth: User not found for sub=%s: %v", sub, err)
		return 0, fmt.Errorf("user not found")
	}

	return userID, nil
}

// lookupOrCreateUser looks up a user by sub and creates them if they don't exist
func lookupOrCreateUser(sub, email string) (int64, error) {
	pool := database.GetPool()
	if pool == nil {
		log.Printf("Auth: Database pool not available")
		return 0, fmt.Errorf("database not connected")
	}

	// First try to find existing user
	var userID int64
	err := pool.QueryRow(context.Background(),
		"SELECT id FROM users WHERE sub = $1", sub).Scan(&userID)
	if err == nil {
		log.Printf("Auth: Found existing user id=%d for sub=%s", userID, sub)
		return userID, nil
	}

	// User doesn't exist, create them
	log.Printf("Auth: Creating new user for sub=%s, email=%s", sub, email)

	err = pool.QueryRow(context.Background(),
		"INSERT INTO users (sub, email) VALUES ($1, $2) RETURNING id",
		sub, email).Scan(&userID)
	if err != nil {
		log.Printf("Auth: Failed to create user for sub=%s: %v", sub, err)
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Auth: Created new user id=%d for sub=%s", userID, sub)
	return userID, nil
}

// introspectToken validates an opaque/JWE token via Zitadel's introspection endpoint
func introspectToken(zitadelDomain, token string) (*introspectionResponse, error) {
	introspectionURL := fmt.Sprintf("https://%s/oauth/v2/introspect", zitadelDomain)

	data := strings.NewReader(fmt.Sprintf("token=%s", token))
	req, err := http.NewRequest("POST", introspectionURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use client credentials if available
	clientID := osGetenv("ZITADEL_CLIENT_ID", "")
	clientSecret := osGetenv("ZITADEL_CLIENT_SECRET", "")
	if clientID != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("introspection request failed: %w", err)
	}
	defer resp.Body.Close()

	var result introspectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode introspection response: %w", err)
	}

	return &result, nil
}

// osGetenv is a helper for getting env vars with defaults
func osGetenv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
