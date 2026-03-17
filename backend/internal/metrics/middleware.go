package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware records HTTP request count and latency for each Gin route.
// Uses c.FullPath() (the route template, e.g. /api/v1/mangas/:mangaID) rather
// than the raw URL so label cardinality stays low regardless of path parameters.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			route,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			route,
		).Observe(time.Since(start).Seconds())
	}
}
