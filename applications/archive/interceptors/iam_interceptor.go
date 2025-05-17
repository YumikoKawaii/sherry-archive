package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"sherry.archive.com/applications/archive/adapters/iam"
)

type IamInterceptor interface {
	Unary(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
}

type iamInterceptorImpl struct {
	iamClient iam.Client
}

func NewIamInterceptor(iamClient iam.Client) IamInterceptor {
	return &iamInterceptorImpl{
		iamClient: iamClient,
	}
}

func (i *iamInterceptorImpl) Unary(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := i.iamClient.Verify(ctx, info.FullMethod); err != nil {
		return nil, err
	}
	return handler(ctx, request)
}
