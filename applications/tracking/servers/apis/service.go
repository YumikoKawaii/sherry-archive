package apis

import (
	"context"
	"github.com/golang/protobuf/proto"
	"net/http"
	pb "sherry.archive.com/pb/tracking"
	"sherry.archive.com/shared/topics"
	"sherry.archive.com/shared/tracking_events"
)

type Service struct {
	pb.UnimplementedTrackingServiceServer
	publisher topics.Publisher
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) LogEntry(ctx context.Context, request *pb.LogEntryRequest) (*pb.LogEntryResponse, error) {
	f, found := tracking_events.Registry[request.Schema]
	if !found {
		return &pb.LogEntryResponse{
			Code:    uint32(http.StatusNotFound),
			Message: "Not found",
		}, nil
	}
	event := f()
	if err := proto.Unmarshal(request.Data, event); err != nil {
		return nil, err
	}

	return nil, nil
}
