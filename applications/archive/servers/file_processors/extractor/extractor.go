package extractor

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"sherry.archive.com/shared/constants"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/topics"
)

type Extractor interface {
	Extract(ctx context.Context)
}

func NewExtractor(consumer topics.Consumer) Extractor {
	return &extractorImpl{
		consumer: consumer,
	}
}

type extractorImpl struct {
	consumer topics.Consumer
}

func (e *extractorImpl) Extract(ctx context.Context) {
	errChan := make(chan error)
	go func() {
		if err := e.consumer.Consume(ctx, constants.MultimediaCompressionTopic, e.extractFile); err != nil {
			errChan <- err
		}
	}()
	select {
	case err := <-errChan:
		logger.Errorf("error consume message on topic %s: %s", constants.MultimediaCompressionTopic, err.Error())
	case <-ctx.Done():
		if err := e.consumer.Close(); err != nil {
			logger.Errorf("error closing consumer: %s:", err.Error())
		}
	}
}

func (e *extractorImpl) extractFile(message proto.Message) error {
	return errors.New("")
}
