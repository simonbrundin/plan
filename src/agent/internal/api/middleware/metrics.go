package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"plan-agent/internal/api/metrics"
)

// Metrics returns a Gin middleware that records HTTP metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		m := metrics.Get()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record start
		m.HTTPRequestStart(path)
		start := time.Now()

		// Process request
		c.Next()

		// Record end
		duration := time.Since(start)
		m.HTTPRequestEnd(path, duration)
	}
}
