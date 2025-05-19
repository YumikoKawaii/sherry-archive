package topics

import (
	"github.com/IBM/sarama"
	"sherry.archive.com/shared/logger"
)

type ConsumerGroupHandler struct {
	handleFn HandleMessageFunc
}

func NewConsumerGroupHandler(handleFn HandleMessageFunc) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		handleFn: handleFn,
	}
}

func (c *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			logger.Infof("received message: timestamp = %s, topic = %s, offset = %d, partition = %d", message.Timestamp, message.Topic, message.Offset, message.Partition)
			if err := c.handleFn(message); err != nil {
				return err
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
