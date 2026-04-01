package middleware

import (
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gin-gonic/gin"
)

// XRay wraps each request in an X-Ray segment named serviceName.
// It must run before any middleware that creates subsegments.
func XRay(serviceName string) gin.HandlerFunc {
	namer := xray.NewFixedSegmentNamer(serviceName)
	return func(c *gin.Context) {
		xray.Handler(namer, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}
