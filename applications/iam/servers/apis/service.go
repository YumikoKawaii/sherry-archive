package apis

import (
	"context"
	"sherry.archive.com/applications/iam/pkg/repository"
	pb "sherry.archive.com/pb/iam"
)

type Service struct {
	*pb.UnimplementedIdentityServiceServer
	querier repository.Querier
}

func NewService(querier repository.Querier) *Service {
	return &Service{
		querier: querier,
	}
}

func (s *Service) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return nil, nil
}

func (s *Service) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	return nil, nil
}

func (s *Service) GetUserDetail(ctx context.Context, request *pb.GetUserDetailRequest) (*pb.GetUserDetailResponse, error) {
	return nil, nil
}

func (s *Service) UpsertUser(ctx context.Context, request *pb.UpsertUserRequest) (*pb.UpsertUserResponse, error) {
	return nil, nil
}

func (s *Service) Verify(ctx context.Context, request *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	return nil, nil
}
