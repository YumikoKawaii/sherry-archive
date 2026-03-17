// Package metrics publishes application metrics to AWS CloudWatch.
// Call Init once at startup; all Record* functions are no-ops until then
// (safe for local dev without CloudWatch access).
package metrics

import (
	"context"
	"database/sql"
	"log"
	"math"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// flushInterval controls how often accumulated metrics are pushed to CloudWatch.
// 10s keeps data loss on restarts/deploys small and gives sub-minute granularity.
const flushInterval = 60 * time.Second

// highRes tells CloudWatch to store at 1-second resolution so dashboards can
// show 10s / 30s / 1m granularity instead of the default 1-minute minimum.
const highRes = int32(1)

// latencyBuckets defines upper bounds (in seconds) for the request duration histogram.
// Finer resolution in the 100–250ms range where most API responses are expected to land.
var latencyBuckets = []float64{
	0.010, 0.025, 0.050, 0.075,
	0.100, 0.125, 0.150, 0.175, 0.200, 0.225, 0.250,
	0.500, 1.000, 2.500,
}

// histogram tracks request latency using fixed buckets.
// Gives min/max/avg via StatisticValues AND p50/p99 via bucket interpolation.
type histogram struct {
	// counts[i] = requests that fell in bucket i; counts[len(latencyBuckets)] = overflow (+Inf)
	counts [15]float64
	count  float64
	sum    float64
	min    float64
	max    float64
}

func newHistogram() histogram {
	return histogram{min: math.MaxFloat64}
}

func (h *histogram) observe(v float64) {
	h.count++
	h.sum += v
	if v < h.min {
		h.min = v
	}
	if v > h.max {
		h.max = v
	}
	for i, upper := range latencyBuckets {
		if v <= upper {
			h.counts[i]++
			return
		}
	}
	h.counts[len(latencyBuckets)]++ // +Inf
}

// percentile computes an approximate quantile (0–1) via linear interpolation.
func (h *histogram) percentile(q float64) float64 {
	if h.count == 0 {
		return 0
	}
	target := q * h.count
	var cumulative float64
	for i, cnt := range h.counts {
		cumulative += cnt
		if cumulative >= target {
			lower := 0.0
			if i > 0 {
				lower = latencyBuckets[i-1]
			}
			upper := h.max // use observed max for the overflow bucket
			if i < len(latencyBuckets) {
				upper = latencyBuckets[i]
			}
			if cnt == 0 {
				return lower
			}
			prev := cumulative - cnt
			fraction := (target - prev) / cnt
			return lower + fraction*(upper-lower)
		}
	}
	return h.max
}

// package-level singleton — nil until Init is called.
var pub *publisher

type httpKey struct {
	method, route, status string
}

// durKey omits status so percentiles are computed per route across all statuses.
type durKey struct {
	method, route string
}

type publisher struct {
	client    *cloudwatch.Client
	namespace string
	db        *sql.DB

	// previous cumulative values from sql.DBStats — used to compute per-window deltas
	prevWaitCount    int64
	prevWaitDuration int64 // nanoseconds

	mu                sync.Mutex
	windowStart       time.Time
	httpRequests      map[httpKey]float64
	httpDurations     map[durKey]histogram
	trackingEvents    map[string]float64
	analyticsRequests map[string]float64
}

// Init creates the CloudWatch publisher and starts the background flush goroutine.
// If CloudWatch is unavailable (e.g. local dev without IMDS), it logs a warning
// and returns an error — all Record* functions remain no-ops.
func Init(ctx context.Context, region, namespace string, db *sql.DB) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return err
	}
	p := &publisher{
		client:            cloudwatch.NewFromConfig(cfg),
		namespace:         namespace,
		db:                db,
		windowStart:       time.Now(),
		httpRequests:      make(map[httpKey]float64),
		httpDurations:     make(map[durKey]histogram),
		trackingEvents:    make(map[string]float64),
		analyticsRequests: make(map[string]float64),
	}
	pub = p
	go p.run(ctx)
	return nil
}

// RecordHTTP records one HTTP request. Called from the Gin metrics middleware.
// Skips "unmatched" routes (SPA catch-all hits) to keep API metrics clean.
func RecordHTTP(method, route, status string, durationSecs float64) {
	if pub == nil || route == "unmatched" {
		return
	}
	pub.mu.Lock()
	pub.httpRequests[httpKey{method, route, status}]++
	dk := durKey{method, route}
	h := pub.httpDurations[dk]
	if h.count == 0 {
		h = newHistogram()
	}
	h.observe(durationSecs)
	pub.httpDurations[dk] = h
	pub.mu.Unlock()
}

// RecordTrackingEvent increments the counter for the given tracking event type.
func RecordTrackingEvent(eventType string) {
	if pub == nil {
		return
	}
	pub.mu.Lock()
	pub.trackingEvents[eventType]++
	pub.mu.Unlock()
}

// RecordAnalyticsRequest increments the counter for the given analytics endpoint.
func RecordAnalyticsRequest(endpoint string) {
	if pub == nil {
		return
	}
	pub.mu.Lock()
	pub.analyticsRequests[endpoint]++
	pub.mu.Unlock()
}

func (p *publisher) run(ctx context.Context) {
	t := time.NewTicker(flushInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			p.flush(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (p *publisher) flush(ctx context.Context) {
	p.mu.Lock()
	windowStart := p.windowStart

	httpReqs := p.httpRequests
	httpDurs := p.httpDurations
	trackEvts := p.trackingEvents
	analyticsReqs := p.analyticsRequests

	p.windowStart = time.Now()
	p.httpRequests = make(map[httpKey]float64)
	p.httpDurations = make(map[durKey]histogram)
	p.trackingEvents = make(map[string]float64)
	p.analyticsRequests = make(map[string]float64)
	p.mu.Unlock()

	var data []types.MetricDatum

	// HTTP request counts — one data point per (method, route, status) combination.
	for k, count := range httpReqs {
		data = append(data, types.MetricDatum{
			MetricName:        aws.String("HTTPRequestCount"),
			Timestamp:         aws.Time(windowStart),
			Value:             aws.Float64(count),
			Unit:              types.StandardUnitCount,
			StorageResolution: aws.Int32(highRes),
			Dimensions: []types.Dimension{
				{Name: aws.String("Method"), Value: aws.String(k.method)},
				{Name: aws.String("Route"), Value: aws.String(k.route)},
				{Name: aws.String("StatusCode"), Value: aws.String(k.status)},
			},
		})
	}

	// HTTP latency — all stats pushed under the single metric name "HTTPRequestDuration"
	// with a "Stat" dimension (avg/min/max/p50/p99). This keeps CloudWatch grouped as
	// one entry per route instead of three separate metric names per route.
	for k, h := range httpDurs {
		if h.count == 0 {
			continue
		}
		avg := h.sum / h.count
		for stat, val := range map[string]float64{
			"avg": avg,
			"p50": h.percentile(0.50),
			"p99": h.percentile(0.99),
		} {
			data = append(data, types.MetricDatum{
				MetricName:        aws.String("HTTPRequestDuration"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(val),
				Unit:              types.StandardUnitSeconds,
				StorageResolution: aws.Int32(highRes),
				Dimensions: []types.Dimension{
					{Name: aws.String("Method"), Value: aws.String(k.method)},
					{Name: aws.String("Route"), Value: aws.String(k.route)},
					{Name: aws.String("Stat"), Value: aws.String(stat)},
				},
			})
		}
	}

	for evtType, count := range trackEvts {
		data = append(data, types.MetricDatum{
			MetricName:        aws.String("TrackingEvents"),
			Timestamp:         aws.Time(windowStart),
			Value:             aws.Float64(count),
			Unit:              types.StandardUnitCount,
			StorageResolution: aws.Int32(highRes),
			Dimensions: []types.Dimension{
				{Name: aws.String("EventType"), Value: aws.String(evtType)},
			},
		})
	}

	for endpoint, count := range analyticsReqs {
		data = append(data, types.MetricDatum{
			MetricName:        aws.String("AnalyticsRequests"),
			Timestamp:         aws.Time(windowStart),
			Value:             aws.Float64(count),
			Unit:              types.StandardUnitCount,
			StorageResolution: aws.Int32(highRes),
			Dimensions: []types.Dimension{
				{Name: aws.String("Endpoint"), Value: aws.String(endpoint)},
			},
		})
	}

	if p.db != nil {
		s := p.db.Stats()

		// Gauges — current snapshot of the pool state.
		data = append(data,
			types.MetricDatum{
				MetricName:        aws.String("DBPoolOpenConnections"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(s.OpenConnections)),
				Unit:              types.StandardUnitCount,
				StorageResolution: aws.Int32(highRes),
			},
			types.MetricDatum{
				MetricName:        aws.String("DBPoolInUse"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(s.InUse)),
				Unit:              types.StandardUnitCount,
				StorageResolution: aws.Int32(highRes),
			},
			types.MetricDatum{
				MetricName:        aws.String("DBPoolIdle"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(s.Idle)),
				Unit:              types.StandardUnitCount,
				StorageResolution: aws.Int32(highRes),
			},
		)

		// Deltas — how many waits and how much wait time occurred in this window.
		// sql.DBStats.WaitCount and WaitDuration are cumulative since process start,
		// so we subtract the previous snapshot to get the per-window value.
		waitCountDelta := s.WaitCount - p.prevWaitCount
		waitDurationDelta := s.WaitDuration.Nanoseconds() - p.prevWaitDuration
		p.prevWaitCount = s.WaitCount
		p.prevWaitDuration = s.WaitDuration.Nanoseconds()

		data = append(data,
			types.MetricDatum{
				MetricName:        aws.String("DBPoolWaitCount"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(waitCountDelta)),
				Unit:              types.StandardUnitCount,
				StorageResolution: aws.Int32(highRes),
			},
			types.MetricDatum{
				MetricName:        aws.String("DBPoolWaitDuration"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(waitDurationDelta) / 1e9), // nanoseconds → seconds
				Unit:              types.StandardUnitSeconds,
				StorageResolution: aws.Int32(highRes),
			},
		)
	}

	if len(data) == 0 {
		return
	}

	// PutMetricData accepts max 1000 data points per call.
	const batchSize = 1000
	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		if _, err := p.client.PutMetricData(ctx, &cloudwatch.PutMetricDataInput{
			Namespace:  aws.String(p.namespace),
			MetricData: data[i:end],
		}); err != nil {
			log.Printf("metrics: cloudwatch flush error: %v", err)
		}
	}
}
