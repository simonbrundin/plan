package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

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
					c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user token"})
					c.Abort()
					return
				}
				userID = userID*10 + int64(ch-'0')
			}
			c.Set("userID", userID)
			c.Set("userSub", fmt.Sprintf("user_%d", userID))
			c.Next()
			return
		}

		// Validate Zitadel JWT
		zitadelDomain := osGetenv("ZITADEL_DOMAIN", "auth.simonbrundin.com")

		// Parse token without validation first to get the kid
		parser := jwt.NewParser()
		token, _, err := parser.ParseUnverified(tokenString, &ZitadelClaims{})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": fmt.Sprintf("Invalid token format: %v", err)})
			c.Abort()
			return
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Token missing kid header"})
			c.Abort()
			return
		}

		// Get the public key for this kid
		publicKey, err := getRSAKey(zitadelDomain, kid)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": fmt.Sprintf("Failed to get public key: %v", err)})
			c.Abort()
			return
		}

		// Parse and validate the token
		claims := &ZitadelClaims{}
		validatedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err != nil || !validatedToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		// Extract user info from claims
		subject := claims.Subject
		email := claims.Email

		// Look up user in database or use default
		userID, err := lookupUserBySub(subject)
		if err != nil {
			// If user not found and AUTH_DISABLED=false, reject
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found in database"})
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
// This should query your database
func lookupUserBySub(sub string) (int64, error) {
	// TODO: Implement database lookup
	// For now, hardcode the known user
	if sub == "378032824856347117" {
		return 1, nil
	}

	// Return error for unknown users
	return 0, fmt.Errorf("user not found")
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
