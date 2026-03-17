// Package metrics publishes application metrics to AWS CloudWatch.
// Call Init once at startup; all Record* functions are no-ops until then
// (safe for local dev without CloudWatch access).
package metrics

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// flushInterval controls how often accumulated metrics are pushed to CloudWatch.
// 10s keeps data loss on restarts/deploys small and gives sub-minute granularity.
const flushInterval = 10 * time.Second

// highRes tells CloudWatch to store at 1-second resolution so dashboards can
// show 10s / 30s / 1m granularity instead of the default 1-minute minimum.
const highRes = int32(1)

// package-level singleton — nil until Init is called.
var pub *publisher

type httpKey struct {
	method, route, status string
}

type durationStats struct {
	count         float64
	sum, min, max float64
}

type publisher struct {
	client    *cloudwatch.Client
	namespace string
	db        *sql.DB

	mu                sync.Mutex
	windowStart       time.Time
	httpRequests      map[httpKey]float64
	httpDurations     map[httpKey]durationStats
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
		httpDurations:     make(map[httpKey]durationStats),
		trackingEvents:    make(map[string]float64),
		analyticsRequests: make(map[string]float64),
	}
	pub = p
	go p.run(ctx)
	return nil
}

// RecordHTTP records one HTTP request. Called from the Gin metrics middleware.
func RecordHTTP(method, route, status string, durationSecs float64) {
	if pub == nil {
		return
	}
	k := httpKey{method, route, status}
	pub.mu.Lock()
	pub.httpRequests[k]++
	s := pub.httpDurations[k]
	s.count++
	s.sum += durationSecs
	if s.count == 1 || durationSecs < s.min {
		s.min = durationSecs
	}
	if durationSecs > s.max {
		s.max = durationSecs
	}
	pub.httpDurations[k] = s
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
	// Capture the window start timestamp before resetting — this is stamped on
	// each data point so CloudWatch knows exactly which 10s window the counts
	// belong to, giving accurate rate calculation (Sum / 10 = req/s).
	windowStart := p.windowStart

	httpReqs := p.httpRequests
	httpDurs := p.httpDurations
	trackEvts := p.trackingEvents
	analyticsReqs := p.analyticsRequests

	p.windowStart = time.Now()
	p.httpRequests = make(map[httpKey]float64)
	p.httpDurations = make(map[httpKey]durationStats)
	p.trackingEvents = make(map[string]float64)
	p.analyticsRequests = make(map[string]float64)
	p.mu.Unlock()

	var data []types.MetricDatum

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

	for k, s := range httpDurs {
		if s.count == 0 {
			continue
		}
		data = append(data, types.MetricDatum{
			MetricName:        aws.String("HTTPRequestDuration"),
			Timestamp:         aws.Time(windowStart),
			Unit:              types.StandardUnitSeconds,
			StorageResolution: aws.Int32(highRes),
			StatisticValues: &types.StatisticSet{
				SampleCount: aws.Float64(s.count),
				Sum:         aws.Float64(s.sum),
				Minimum:     aws.Float64(s.min),
				Maximum:     aws.Float64(s.max),
			},
			Dimensions: []types.Dimension{
				{Name: aws.String("Method"), Value: aws.String(k.method)},
				{Name: aws.String("Route"), Value: aws.String(k.route)},
			},
		})
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
		data = append(data,
			types.MetricDatum{
				MetricName:        aws.String("DBOpenConnections"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(s.OpenConnections)),
				Unit:              types.StandardUnitCount,
				StorageResolution: aws.Int32(highRes),
			},
			types.MetricDatum{
				MetricName:        aws.String("DBInUseConnections"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(s.InUse)),
				Unit:              types.StandardUnitCount,
				StorageResolution: aws.Int32(highRes),
			},
			types.MetricDatum{
				MetricName:        aws.String("DBIdleConnections"),
				Timestamp:         aws.Time(windowStart),
				Value:             aws.Float64(float64(s.Idle)),
				Unit:              types.StandardUnitCount,
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
