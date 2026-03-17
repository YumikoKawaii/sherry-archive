package metrics

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP layer
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests partitioned by method, route template, and status code.",
		},
		[]string{"method", "route", "status_code"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency partitioned by method and route template.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	// Database connection pool
	DBOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_open_connections",
		Help: "Current number of open database connections (in-use + idle).",
	})

	DBInUseConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_in_use_connections",
		Help: "Current number of database connections in use.",
	})

	DBIdleConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_idle_connections",
		Help: "Current number of idle database connections in the pool.",
	})

	DBWaitCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_wait_count_total",
		Help: "Cumulative number of times a goroutine waited for a database connection.",
	})

	// Business metrics
	TrackingEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tracking_events_total",
			Help: "Total tracking events ingested, partitioned by event type.",
		},
		[]string{"event_type"},
	)

	AnalyticsRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_requests_total",
			Help: "Total analytics endpoint requests, partitioned by endpoint.",
		},
		[]string{"endpoint"},
	)
)

func init() {
	prometheus.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		HTTPRequestsTotal,
		HTTPRequestDuration,
		DBOpenConnections,
		DBInUseConnections,
		DBIdleConnections,
		DBWaitCount,
		TrackingEventsTotal,
		AnalyticsRequestsTotal,
	)
}

// Handler returns the Prometheus HTTP handler to mount at /metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}

// CollectDBStats polls db.Stats() every 15s and updates the DB pool gauges.
// Runs until ctx is cancelled (call from a background goroutine).
func CollectDBStats(ctx context.Context, db *sql.DB) {
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s := db.Stats()
			DBOpenConnections.Set(float64(s.OpenConnections))
			DBInUseConnections.Set(float64(s.InUse))
			DBIdleConnections.Set(float64(s.Idle))
			DBWaitCount.Set(float64(s.WaitCount))
		case <-ctx.Done():
			return
		}
	}
}
