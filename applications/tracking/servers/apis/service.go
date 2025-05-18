package apis

import (
	"context"
	pb "sherry.archive.com/pb/tracking"
	"sherry.archive.com/shared/topics"
)

type Service struct {
	pb.UnimplementedTrackingServiceServer
	publisher topics.Publisher
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) LogEntry(ctx context.Context, request *pb.LogEntryRequest) (*pb.LogEntryResponse, error) {
	return nil, nil
}
