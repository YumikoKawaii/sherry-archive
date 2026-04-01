// Package tracing configures AWS X-Ray for distributed tracing.
// Call Init once at startup; all instrumentation is a no-op when disabled.
package tracing

import (
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/yumikokawaii/sherry-archive/internal/config"
)

// Init configures the X-Ray SDK. When disabled, the SDK is not initialised
// and all xray.BeginSubsegment / xray.Capture calls are safe no-ops.
func Init(cfg *config.TracingConfig) error {
	if !cfg.Enabled {
		return nil
	}
	return xray.Configure(xray.Config{
		DaemonAddr:     cfg.DaemonAddr,
		ServiceVersion: "1.0.0",
	})
}
