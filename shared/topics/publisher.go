package topics

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"sherry.archive.com/shared/configs"
	"sync"
)

type Publisher interface {
	Publish(ctx context.Context, message proto.Message, topic string, key *string) error
}

func NewKafkaSyncPublisher(cfg configs.KafkaConfig) Publisher {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	//config.Producer.Idempotent = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		panic(err)
	}

	return &kafkaSyncPublisher{
		publisher: producer,
	}
}

type kafkaSyncPublisher struct {
	publisher sarama.SyncProducer
	mutex     sync.Mutex
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
	return err
}
