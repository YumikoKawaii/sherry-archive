package topics

import (
	"context"
	"errors"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"sherry.archive.com/shared/configs"
	"sherry.archive.com/shared/logger"
)

type KafkaProducerTypeOption string

const (
	SyncKafka  KafkaProducerTypeOption = "Sync"
	AsyncKafka KafkaProducerTypeOption = "Async"
)

type Publisher interface {
	Publish(ctx context.Context, message proto.Message, topic string, key *string) error
	PublishInBatch(ctx context.Context, messages []proto.Message, topic string, key *string) error
}

func NewKafkaSyncPublisher(cfg configs.KafkaConfig, producerType KafkaProducerTypeOption) Publisher {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	switch producerType {
	case SyncKafka:
		producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
		if err != nil {
			panic(err)
		}

		return &kafkaSyncPublisher{
			publisher: producer,
		}
	default:
		producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
		if err != nil {
			panic(err)
		}

		return &kafkaAsyncProducer{
			publisher: producer,
		}
	}
}

type kafkaSyncPublisher struct {
	publisher sarama.SyncProducer
}

func (p *kafkaSyncPublisher) Publish(ctx context.Context, message proto.Message, topic string, key *string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	var msgKey sarama.Encoder
	if key != nil {
		msgKey = sarama.StringEncoder(*key)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   msgKey,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.publisher.SendMessage(msg)
	if err != nil {
		logger.Infof("publish message to topic %s", topic)
	}
	return err
}

func (p *kafkaSyncPublisher) PublishInBatch(ctx context.Context, messages []proto.Message, topic string, key *string) error {
	return errors.New("unsupported")
}

type kafkaAsyncProducer struct {
	publisher sarama.AsyncProducer
}

func (p *kafkaAsyncProducer) Publish(ctx context.Context, message proto.Message, topic string, key *string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	var msgKey sarama.Encoder
	if key != nil {
		msgKey = sarama.StringEncoder(*key)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   msgKey,
		Value: sarama.ByteEncoder(data),
	}

	p.publisher.Input() <- msg

	return nil
}

func (p *kafkaAsyncProducer) PublishInBatch(ctx context.Context, messages []proto.Message, topic string, key *string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	for _, message := range messages {
		data, err := proto.Marshal(message)
		if err != nil {
			logger.Errorf("error marshaling message: %s", err.Error())
			continue
		}
		var msgKey sarama.Encoder
		if key != nil {
			msgKey = sarama.StringEncoder(*key)
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Key:   msgKey,
			Value: sarama.ByteEncoder(data),
		}

		p.publisher.Input() <- msg
	}

	return nil
}
