package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware records HTTP request count and latency for each Gin route.
// Uses c.FullPath() (the route template, e.g. /api/v1/mangas/:mangaID) rather
// than the raw URL so dimension cardinality stays low regardless of path parameters.
// Only routes under /api/v1 are recorded; static assets and other paths are skipped.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/v1") {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		RecordHTTP(
			c.Request.Method,
			route,
			strconv.Itoa(c.Writer.Status()),
			time.Since(start).Seconds(),
		)
	}
}
