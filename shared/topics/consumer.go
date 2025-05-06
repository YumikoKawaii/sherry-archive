package topics

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"sherry.archive.com/shared/configs"
	"sherry.archive.com/shared/logger"
	"sync"
	"time"
)

type Consumer interface {
	Consume(ctx context.Context, topic string, handler func(message *sarama.ConsumerMessage) error) error
	Close() error
}

func NewKafkaConsumer(cfg configs.KafkaConfig) Consumer {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.MaxProcessingTime = 500 * time.Millisecond

	config.Consumer.Fetch.Default = 1024 * 1024  // 1MB
	config.Consumer.Fetch.Max = 10 * 1024 * 1024 // 10MB

	// Create client
	client, err := sarama.NewClient(cfg.Brokers, config)
	if err != nil {
		panic(err)
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		client.Close()
		panic(err)
	}

	return &kafkaConsumer{
		consumer: consumer,
		client:   client,
	}
}

type kafkaConsumer struct {
	consumer sarama.Consumer
	client   sarama.Client
	//groupId  string
}

func (c *kafkaConsumer) Consume(ctx context.Context, topic string, handler func(message *sarama.ConsumerMessage) error) error {
	partitions, err := c.consumer.Partitions(topic)
	if err != nil {
		return err
	}

	if len(partitions) == 0 {
		return errors.New("no available partitions on topic")
	}

	var wg sync.WaitGroup
	errChan := make(chan error)
	for _, partition := range partitions {
		wg.Add(1)

		go func(partition int32) {
			defer wg.Done()
			consumer, err := c.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
			if err != nil {
				errChan <- err
				return
			}
			defer consumer.Close()

			for {
				select {
				case msg := <-consumer.Messages():
					logger.Infof("receive message on topic %s, partition %d", topic, partition)
					if err := handler(msg); err != nil {
						logger.Errorf("error process message on topic %s: %s", topic, err.Error())
						errChan <- err
						return
					}
				case err := <-consumer.Errors():
					logger.Errorf("error consume message on topic %s: %s", topic, err.Error())
				case <-ctx.Done():
					return
				}
			}
		}(partition)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *kafkaConsumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return err
	}

	return c.client.Close()
}
