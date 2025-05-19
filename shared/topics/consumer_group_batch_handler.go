package topics

import (
	"github.com/IBM/sarama"
	"sherry.archive.com/shared/logger"
	"time"
)

type ConsumerGroupBatchHandler struct {
	batchSize int
	cycle     time.Duration
	handlerFn HandleBatchMessageFunc
}

func NewConsumerGroupBatchHandler(batchSize int, cycle time.Duration, handleFn HandleBatchMessageFunc) *ConsumerGroupBatchHandler {
	return &ConsumerGroupBatchHandler{
		batchSize: batchSize,
		cycle:     cycle,
		handlerFn: handleFn,
	}
}

func (c *ConsumerGroupBatchHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerGroupBatchHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerGroupBatchHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	messages := make([]*sarama.ConsumerMessage, 0)
	ticker := time.NewTicker(c.cycle)
	defer ticker.Stop()
	for {
		select {
		case message := <-claim.Messages():
			logger.Infof("received message: timestamp = %s, topic = %s, offset = %d, partition = %d", message.Timestamp, message.Topic, message.Offset, message.Partition)
			messages = append(messages, message)
			if len(messages) < c.batchSize {
				continue
			}

			if err := c.handlerFn(messages); err != nil {
				return err
			}
			messages = make([]*sarama.ConsumerMessage, 0)
			session.MarkMessage(message, "")
		case <-ticker.C:
			if len(messages) == 0 {
				continue
			}

			if err := c.handlerFn(messages); err != nil {
				return err
			}
			session.MarkMessage(messages[len(messages)-1], "")
			messages = make([]*sarama.ConsumerMessage, 0)
		case <-session.Context().Done():
			return nil
		}
	}
}
