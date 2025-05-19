package apis

import (
	"context"
	"github.com/golang/protobuf/proto"
	"net/http"
	"sherry.archive.com/applications/tracking/pkg/constants"
	pb "sherry.archive.com/pb/tracking"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/topics"
)

type Service struct {
	pb.UnimplementedTrackingServiceServer
	publisher topics.Publisher
}

func NewService(publisher topics.Publisher) *Service {
	return &Service{
		publisher: publisher,
	}
}

func (s *Service) LogEntry(ctx context.Context, request *pb.LogEntryRequest) (*pb.LogEntryResponse, error) {
	err := s.publisher.Publish(ctx, request.LogEntry, constants.LogEntriesTopic, nil)
	if err != nil {
		return nil, err
	}

	return &pb.LogEntryResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
	}, nil
}

func (s *Service) LogEntries(ctx context.Context, request *pb.LogEntriesRequest) (*pb.LogEntriesResponse, error) {
	messages := make([]proto.Message, 0)
	for _, log := range request.LogEntries {
		messages = append(messages, log)
	}
	if err := s.publisher.PublishInBatch(ctx, messages, constants.LogEntriesTopic, nil); err != nil {
		logger.Errorf("error publish batch: %s", err.Error())
	}
	return &pb.LogEntriesResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
	}, nil
}
