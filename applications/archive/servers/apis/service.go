package apis

import (
	"context"
	"google.golang.org/grpc/codes"
	pb "sherry.archive.com/pb/archive"
)

type Service struct {
	pb.ArchiveServiceServer
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Dummy(ctx context.Context, request *pb.DummyRequest) (*pb.DummyResponse, error) {
	return &pb.DummyResponse{
		Code: int32(codes.OK),
	}, nil
}
