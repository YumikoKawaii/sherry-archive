package topics

import (
	"context"
	"github.com/IBM/sarama"
	"sherry.archive.com/shared/configs"
	"sherry.archive.com/shared/logger"
	"strings"
	"time"
)

type consumerGroup struct {
	group     sarama.ConsumerGroup
	closeChan chan struct{}
	batchSize int
	cycle     time.Duration
}

func NewKafkaConsumerGroup(cfg configs.KafkaConfig) Consumer {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupId, saramaCfg)
	if err != nil {
		panic(err)
	}
	return &consumerGroup{
		group:     group,
		closeChan: make(chan struct{}),
		batchSize: cfg.BatchSize,
		cycle:     time.Duration(cfg.CycleInSec) * time.Second,
	}
}

func (c *consumerGroup) ConsumeInBatch(ctx context.Context, topic string, fn HandleBatchMessageFunc) error {
	errChan := make(chan error)
	handler := NewConsumerGroupBatchHandler(c.batchSize, c.cycle, fn)
	go func(errChan chan error) {
		for {
			select {
			case <-c.closeChan:
				logger.Info("closing consumer group")
				return
			default:
				if err := c.group.Consume(ctx, strings.Split(topic, ","), handler); err != nil {
					errChan <- err
				}

				if ctx.Err() != nil {
					logger.Errorf("context was canceled: %s", ctx.Err())
					return
				}
			}
		}
	}(errChan)
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *consumerGroup) Consume(ctx context.Context, topic string, fn HandleMessageFunc) error {
	errChan := make(chan error)
	handler := NewConsumerGroupHandler(fn)
	go func(errChan chan error) {
		for {
			select {
			case <-c.closeChan:
				logger.Info("closing consumer group")
				return
			default:
				if err := c.group.Consume(ctx, strings.Split(topic, ","), handler); err != nil {
					errChan <- err
				}

				if ctx.Err() != nil {
					logger.Errorf("context was canceled: %s", ctx.Err())
					return
				}
			}
		}
	}(errChan)
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *consumerGroup) Close() error {
	return c.group.Close()
}
