package extractor

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/go-redis/redis/v8"
	"github.com/golang/protobuf/proto"
	"gorm.io/gorm"
	"sherry.archive.com/applications/tracking/pkg/constants"
	"sherry.archive.com/applications/tracking/pkg/repository"
	pb "sherry.archive.com/pb/tracking"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/topics"
	"sherry.archive.com/shared/tracking_events"
	"time"
)

type Batch struct {
	Schema string
	Events []proto.Message
}

type Extractor interface {
	Extract(ctx context.Context)
}

func NewExtractor(consumer topics.Consumer, publisher topics.Publisher, querier repository.Querier, redisClient *redis.Client, cacheTTLInSec int) Extractor {
	return &extractorImpl{
		consumer:      consumer,
		publisher:     publisher,
		querier:       querier,
		redisClient:   redisClient,
		cacheTTLInSec: cacheTTLInSec,
	}
}

type extractorImpl struct {
	consumer      topics.Consumer
	publisher     topics.Publisher
	querier       repository.Querier
	redisClient   *redis.Client
	cacheTTLInSec int
}

func (e *extractorImpl) Extract(ctx context.Context) {
	// filter events
	errChan := make(chan error)
	go func() {
		if err := e.consumer.ConsumeInBatch(ctx, constants.LogEntriesTopic, e.consumeMessagesInBatch); err != nil {
			errChan <- err
		}
	}()
	select {
	case err := <-errChan:
		logger.Errorf("error consume message on topic %s: %s", constants.LogEntriesTopic, err.Error())
	case <-ctx.Done():
		if err := e.consumer.Close(); err != nil {
			logger.Errorf("error closing consumer: %s:", err.Error())
		}
	}
}

func (e *extractorImpl) consumeMessagesInBatch(messages []*sarama.ConsumerMessage) error {
	eventsBySchema := make(map[string][]proto.Message)
	for _, message := range messages {
		logEntry := &pb.LogEntry{}
		if err := proto.Unmarshal(message.Value, logEntry); err != nil {
			logger.Errorf("error unmarshaling log entry: %s", err.Error())
			continue
		}

		f, ok := tracking_events.Registry[logEntry.Schema]
		if !ok {
			logger.Errorf("schema not found: %s", logEntry.Schema)
			continue
		}
		event := f()
		if err := proto.Unmarshal(logEntry.Payload, event); err != nil {
			logger.Errorf("error unmarshaling %s: %s", logEntry.Schema, err.Error())
			continue
		}

		// filter events
		trackingId := event.GetTrackingId()
		valid, err := e.redisClient.Get(context.Background(), trackingId).Bool()
		if err != nil {
			logger.Errorf("error fetching data from redis: %s", err.Error())
			_, err := e.querier.GetTrackingId(context.Background(), trackingId)
			if err != nil {
				switch err {
				case gorm.ErrRecordNotFound:
					if err := e.redisClient.Set(context.Background(), trackingId, false, time.Duration(e.cacheTTLInSec)*time.Second).Err(); err != nil {
						logger.Errorf("error setting data to redis: %s", err.Error())
					}
					valid = false
				default:
					logger.Errorf("error getting data from database: %s", err.Error())
				}
			} else {
				if err := e.redisClient.Set(context.Background(), trackingId, true, time.Duration(e.cacheTTLInSec)*time.Second).Err(); err != nil {
					logger.Errorf("error setting data to redis: %s", err.Error())
				}
				valid = true
			}
		}

		if valid {
			events, found := eventsBySchema[logEntry.Schema]
			if !found {
				events = make([]proto.Message, 0)
			}
			events = append(events, event)
			eventsBySchema[logEntry.Schema] = events
		}
	}
	// splits events into batches
	batches := make([]Batch, 0)
	for schema, events := range eventsBySchema {
		batches = append(batches, Batch{
			Schema: schema,
			Events: events,
		})
	}
	// publish events into it topic
	for _, batch := range batches {
		if err := e.publisher.PublishInBatch(context.Background(), batch.Events, batch.Schema, nil); err != nil {
			logger.Errorf("error publishing events: %s", err.Error())
		}
	}
	return nil
}
