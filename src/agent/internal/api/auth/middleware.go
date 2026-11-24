package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// BearerSchema is the auth schema name
	BearerSchema = "BearerAuth"
	// ContextUserKey is the context key for user claims
	ContextUserKey = "user_claims"
)

// AuthMiddleware creates a middleware for JWT authentication
func AuthMiddleware(config *Config) gin.HandlerFunc {
	validator := NewValidator(config)

	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "empty token",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		claims, err := validator.ValidateAccessToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			message := "invalid token"
			code := "UNAUTHORIZED"

			if err == ErrExpiredToken {
				message = "token has expired"
				code = "TOKEN_EXPIRED"
			}

			c.AbortWithStatusJSON(status, gin.H{
				"error": message,
				"code":  code,
			})
			return
		}

		// Store claims in context
		c.Set(ContextUserKey, claims)
		c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that optionally authenticates
func OptionalAuthMiddleware(config *Config) gin.HandlerFunc {
	validator := NewValidator(config)

	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		if tokenString == "" {
			c.Next()
			return
		}

		claims, err := validator.ValidateAccessToken(tokenString)
		if err == nil {
			c.Set(ContextUserKey, claims)
		}

		c.Next()
	}
}

// RoleMiddleware creates a middleware that checks for required roles
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := GetUserClaims(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		if len(requiredRoles) > 0 && !claims.HasAnyRole(requiredRoles...) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":          "insufficient permissions",
				"code":           "FORBIDDEN",
				"required_roles": requiredRoles,
			})
			return
		}

		c.Next()
	}
}

// GetUserClaims extracts user claims from the gin context
func GetUserClaims(c *gin.Context) (*Claims, bool) {
	value, exists := c.Get(ContextUserKey)
	if !exists {
		return nil, false
	}

	claims, ok := value.(*Claims)
	return claims, ok
}

// GetUserID extracts the user ID from the gin context
func GetUserID(c *gin.Context) (string, bool) {
	claims, ok := GetUserClaims(c)
	if !ok {
		return "", false
	}
	return claims.UserID, true
}

// RequireUserID is a helper that extracts and validates user ID
func RequireUserID(c *gin.Context) (string, error) {
	userID, ok := GetUserID(c)
	if !ok {
		return "", ErrInvalidClaims
	}
	if userID == "" {
		return "", ErrInvalidClaims
	}
	return userID, nil
}

// AdminOnly middleware restricts access to admin users
func AdminOnly() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// UserOwnerOrAdmin middleware checks if user owns the resource or is admin
func UserOwnerOrAdmin(getResourceUserID func(c *gin.Context) (string, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetUserClaims(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Admins can access everything
		if claims.HasRole("admin") {
			c.Next()
			return
		}

		// Check if user owns the resource
		resourceUserID, err := getResourceUserID(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "access denied",
				"code":  "FORBIDDEN",
			})
			return
		}

		if claims.UserID != resourceUserID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "access denied",
				"code":  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}
