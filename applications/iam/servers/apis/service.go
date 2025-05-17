package apis

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/metadata"
	"net/http"
	"sherry.archive.com/applications/iam/pkg/constants"
	"sherry.archive.com/applications/iam/pkg/jwt"
	"sherry.archive.com/applications/iam/pkg/repository"
	pb "sherry.archive.com/pb/iam"
	"sherry.archive.com/shared/hash"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/proto_values"
)

type Config struct {
	HashKey []byte
}

type Service struct {
	*pb.UnimplementedIdentityServiceServer
	cfg         *Config
	querier     repository.Querier
	jwtResolver jwt.Resolver
	redisClient *redis.Client
}

func NewService(cfg *Config, querier repository.Querier, jwtResolver jwt.Resolver, redisClient *redis.Client) *Service {
	return &Service{
		cfg:         cfg,
		querier:     querier,
		jwtResolver: jwtResolver,
		redisClient: redisClient,
	}
}

func (s *Service) Register(ctx context.Context, request *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	passwordData, err := hex.DecodeString(request.Password)
	if err != nil {
		logger.Errorf("error decoding password to data: %s", err.Error())
		return nil, err
	}

	hashedPassword, err := hash.Blake2b256(s.cfg.HashKey, passwordData)
	if err != nil {
		logger.Errorf("error hashing password: %s", err.Error())
		return nil, err
	}

	// TODO: verify email here
	user := &repository.User{
		Email:          request.Email,
		HashedPassword: hex.EncodeToString(hashedPassword),
		Status:         constants.ReaderRole,
	}
	if err := s.querier.InitialUser(ctx, user); err != nil {
		logger.Errorf("error upserting account: %s", err.Error())
		return nil, err
	}

	// generate token
	token, err := s.jwtResolver.GenerateToken(jwt.TokenParameters{
		UserId: user.Id,
		Roles:  []string{constants.ReaderRole},
	})

	if err != nil {
		logger.Errorf("error generating jwt token: %s", err.Error())
		return nil, err
	}

	return &pb.RegisterResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data: &pb.RegisterResponse_Data{
			Token: token,
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	passwordData, err := hex.DecodeString(request.Password)
	if err != nil {
		logger.Errorf("error decoding password to data: %s", err.Error())
		return nil, err
	}

	hashedPassword, err := hash.Blake2b256(s.cfg.HashKey, passwordData)
	if err != nil {
		logger.Errorf("error hashing password: %s", err.Error())
		return nil, err
	}
	// find user
	users, err := s.querier.GetUsers(ctx, &repository.GetUsersFilter{
		Status:         proto_values.StringToPointer(constants.ActiveStatus),
		Email:          &request.Email,
		HashedPassword: proto_values.StringToPointer(hex.EncodeToString(hashedPassword)),
	})
	if err != nil {
		logger.Errorf("error finding users: %s", err.Error())
		return nil, err
	}

	if len(users) != 1 {
		return &pb.LoginResponse{
			Code:    uint32(http.StatusBadRequest),
			Message: "Invalid",
		}, nil
	}
	user := users[1]
	roles := make([]string, 0)
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	// generate token
	token, err := s.jwtResolver.GenerateToken(jwt.TokenParameters{
		UserId: user.Id,
		Roles:  roles,
	})

	if err != nil {
		logger.Errorf("error generating jwt token: %s", err.Error())
		return nil, err
	}

	return &pb.LoginResponse{
		Code:    uint32(http.StatusOK),
		Message: "Success",
		Data: &pb.LoginResponse_Data{
			Token: token,
		},
	}, nil
}

func (s *Service) GetUserDetail(ctx context.Context, request *pb.GetUserDetailRequest) (*pb.GetUserDetailResponse, error) {
	return nil, nil
}

func (s *Service) UpsertUser(ctx context.Context, request *pb.UpsertUserRequest) (*pb.UpsertUserResponse, error) {
	return nil, nil
}

func (s *Service) Verify(ctx context.Context, request *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &pb.VerifyResponse{
			Code:    uint32(http.StatusBadRequest),
			Message: "Invalid",
		}, nil
	}
	authHeaders := md.Get(constants.AuthorizationHeader)
	if len(authHeaders) == 0 {
		return &pb.VerifyResponse{
			Code:    uint32(http.StatusUnauthorized),
			Message: "Invalid",
		}, nil
	}
	token := authHeaders[0]
	claims, err := s.jwtResolver.ResolveToken(token)
	if err != nil {
		return &pb.VerifyResponse{
			Code:    uint32(http.StatusBadRequest),
			Message: err.Error(),
		}, nil
	}

	for _, role := range claims.Roles {
		key := fmt.Sprintf("%s.%s", constants.RolePrefixCacheKey, role)
		ok, err := s.redisClient.SIsMember(ctx, key, request.Path).Result()
		if err != nil {
			logger.Errorf("error get cache from redis: %s", err.Error())
		}

		if ok {
			return &pb.VerifyResponse{
				Code:    uint32(http.StatusOK),
				Message: "Success",
			}, nil
		}
	}

	return &pb.VerifyResponse{
		Code:    uint32(http.StatusUnauthorized),
		Message: "Unauthorized",
	}, nil
}
