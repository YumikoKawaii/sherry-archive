package middleware

import (
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a Gin middleware that emits a structured JSON access log
// per request, including the X-Ray trace_id for correlation.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("latency_ms", time.Since(start).Milliseconds()),
			zap.String("ip", c.ClientIP()),
		}
		if seg := xray.GetSegment(c.Request.Context()); seg != nil {
			fields = append(fields, zap.String("trace_id", seg.TraceID))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		status := c.Writer.Status()
		switch {
		case status >= 500:
			zap.L().Error("request", fields...)
		case status >= 400:
			zap.L().Warn("request", fields...)
		default:
			zap.L().Info("request", fields...)
		}
	}
}
