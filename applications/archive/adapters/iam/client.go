package iam

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	pb "sherry.archive.com/pb/iam"
)

type Client interface {
	Verify(ctx context.Context, path string) error
}

type clientImpl struct {
	client pb.IdentityServiceClient
}

func NewClient(host string) Client {
	cc, err := grpc.NewClient(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return &clientImpl{
		client: pb.NewIdentityServiceClient(cc),
	}
}

func (c *clientImpl) Verify(ctx context.Context, path string) error {
	resp, err := c.client.Verify(ctx, &pb.VerifyRequest{
		Path: path,
	})
	if err != nil {
		return err
	}

	if resp.Code != http.StatusOK {
		return errors.New(resp.Message)
	}

	return nil
}
