package uploader

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"sherry.archive.com/applications/archive/adapters/multimedia"
	"sherry.archive.com/applications/archive/pkg/repository"
	"sherry.archive.com/pb/messages"
	"sherry.archive.com/shared/constants"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/topics"
)

type Uploader interface {
	Process(ctx context.Context)
}

func NewUploader(storage multimedia.StorageClient, consumer topics.Consumer, querier repository.Querier) Uploader {
	return &uploaderImpl{
		storage:  storage,
		consumer: consumer,
		querier:  querier,
	}
}

type uploaderImpl struct {
	storage  multimedia.StorageClient
	consumer topics.Consumer
	querier  repository.Querier
}

func (u *uploaderImpl) Process(ctx context.Context) {
	errChan := make(chan error)
	go func() {
		if err := u.consumer.Consume(ctx, constants.MultimediaTopic, u.upload); err != nil {
			errChan <- err
		}
	}()
	select {
	case err := <-errChan:
		logger.Errorf("error consume message on topic %s: %s", constants.MultimediaCompressionTopic, err.Error())
	case <-ctx.Done():
		if err := u.consumer.Close(); err != nil {
			logger.Errorf("error closing consumer: %s:", err.Error())
		}
	}
}

func (u *uploaderImpl) upload(message *sarama.ConsumerMessage) error {
	var pageMessage messages.Page
	if err := proto.Unmarshal(message.Value, &pageMessage); err != nil {
		return err
	}

	// upload
	url, err := u.storage.Save(context.Background(), pageMessage.Data)
	if err != nil {
		return err
	}

	// save to database
	page := &repository.Page{
		ChapterId: pageMessage.ChapterId,
		ImageUrl:  url,
		Index:     pageMessage.Index,
	}
	err = u.querier.UpsertPage(context.Background(), page)
	if err != nil {
		logger.Infof("process page %d, chapter_id %d", page.Index, page.ChapterId)
	}
	return err
}
