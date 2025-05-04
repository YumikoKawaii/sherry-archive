package apis

import (
	"context"
	"google.golang.org/grpc/codes"
	"sherry.archive.com/applications/archive/pkg/repository"
	pb "sherry.archive.com/pb/archive"
)

type Service struct {
	pb.ArchiveServiceServer
	querier *repository.Querier
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetBooks(ctx context.Context, request *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	return &pb.GetBookResponse{
		Code:    int32(codes.OK),
		Message: "Success",
		Data:    nil,
	}, nil
}

func (s *Service) GetPages(ctx context.Context, request *pb.GetPagesRequest) (*pb.GetPagesResponse, error) {
	return &pb.GetPagesResponse{
		Code:    int32(codes.OK),
		Message: "Success",
		Data:    nil,
	}, nil
}

func (s *Service) GetAuthors(ctx context.Context, request *pb.GetAuthorsRequest) (*pb.GetAuthorsResponse, error) {
	return &pb.GetAuthorsResponse{
		Code:    int32(codes.OK),
		Message: "Success",
		Data:    nil,
	}, nil
}

func (s *Service) GetPublishers(ctx context.Context, request *pb.GetPublishersRequest) (*pb.GetPublishersResponse, error) {
	return &pb.GetPublishersResponse{
		Code:    int32(codes.OK),
		Message: "Success",
		Data:    nil,
	}, nil
}
