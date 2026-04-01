package middleware

import (
	"net/http"
	"strings"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gin-gonic/gin"
)

// XRay wraps each request in an X-Ray segment named serviceName.
// It must run before any middleware that creates subsegments.
// Static asset requests are passed through without tracing to avoid
// "segment cannot be found" noise from the X-Ray SDK.
func XRay(serviceName string) gin.HandlerFunc {
	namer := xray.NewFixedSegmentNamer(serviceName)
	return func(c *gin.Context) {
		if isStaticPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		xray.Handler(namer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}

func isStaticPath(path string) bool {
	return strings.HasPrefix(path, "/assets/") ||
		path == "/favicon.ico" ||
		path == "/favicon.png" ||
		path == "/apple-touch-icon.png"
}
