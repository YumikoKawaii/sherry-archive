package extractor

import (
	"archive/zip"
	"bytes"
	"context"
	"github.com/IBM/sarama"
	"github.com/golang/protobuf/proto"
	"io"
	"sherry.archive.com/pb/messages"
	"sherry.archive.com/shared/constants"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/topics"
	"sort"
	"strings"
)

type Extractor interface {
	Extract(ctx context.Context)
}

func NewExtractor(consumer topics.Consumer, publisher topics.Publisher) Extractor {
	return &extractorImpl{
		consumer:  consumer,
		publisher: publisher,
	}
}

type extractorImpl struct {
	consumer  topics.Consumer
	publisher topics.Publisher
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

func (e *extractorImpl) extractFile(message *sarama.ConsumerMessage) error {
	var pagesMessage messages.Pages
	if err := proto.Unmarshal(message.Value, &pagesMessage); err != nil {
		return err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(pagesMessage.Data), int64(len(pagesMessage.Data)))
	if err != nil {
		return err
	}
	files := make([]*zip.File, 0)
	for _, file := range zipReader.File {
		if constants.SupportedMultimediaTypes[extractFileExtensionFromName(file.Name)] {
			files = append(files, file)
		}
	}
	// sort the file
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	pages := make([]*messages.Page, 0)
	// create page messages
	for idx, file := range files {
		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		fileBytes, err := io.ReadAll(fileReader)
		if err != nil {
			return err
		}
		pages = append(pages, &messages.Page{
			ChapterId: pagesMessage.ChapterId,
			Index:     uint32(idx),
			Data:      fileBytes,
		})
	}

	for _, page := range pages {
		if err := e.publisher.Publish(context.Background(), page, constants.MultimediaTopic, nil); err != nil {
			return err
		}
	}
	logger.Infof("process pages for book_id %d", pagesMessage.ChapterId)
	return nil
}

func extractFileExtensionFromName(fileName string) string {
	fileNameComponents := strings.Split(fileName, constants.FileNameDelimiter)
	return fileNameComponents[len(fileNameComponents)-1]
}
