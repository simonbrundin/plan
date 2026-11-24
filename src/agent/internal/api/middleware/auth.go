package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// BearerHeader is the header the API key is sent in.
const BearerHeader = "Authorization"

// RequireAPIKey kräver en giltig nyckel på alla anrop.
//
// Nyckeln sätts i AGENT_API_KEY. Utan den öppna nyckeln är tjänsten öppen
// för alla som når den - och då kan någon besluta om ett godkännande åt
// människor. därför loggas det tydligt när tjänsten startas utan nyckel.
//
// Hälsokontrollen är undantagen, den säger ingenting om vad tjänsten får.
func RequireAPIKey() gin.HandlerFunc {
	expected := os.Getenv("AGENT_API_KEY")

	return func(c *gin.Context) {
		if expected == "" {
			c.Next()
			return
		}

		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		if !matchesKey(c.GetHeader(BearerHeader), expected) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "En giltig API-nyckel krävs",
			})
			return
		}

		c.Next()
	}
}

// isPublicPath anger vilka anrop som är öppna utan nyckel.
//
// Det är bara hälsokontrollen - den behöver fungera utan nyckel så att
// drift kan se att tjänsten lever.
func isPublicPath(path string) bool {
	switch path {
	case "/api/v1/health", "/api/v1/status", "/health":
		return true
	default:
		return false
	}
}

// matchesKey jämför den mottagna nyckeln med den förväntade.
//
// Nyckeln ska komma med Bearer-prefix, så att en av misstag skickad
// nyckel inte råkar passera.
func matchesKey(header, expected string) bool {
	token, found := strings.CutPrefix(strings.TrimSpace(header), "Bearer ")
	if !found {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(token)), []byte(expected)) == 1
}
