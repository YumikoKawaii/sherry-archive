package grpc_error

import (
	"context"
	"errors"
	"sherry.archive.com/shared/logger"

	//nolint: staticcheck
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a new unary server interceptor that wraps output error.
// `hideServerError` is to hide internal server error information (stacktrace, error message...). Must be true in production.
// Output error will be converted to GRPC error before sending to clients.
func UnaryServerInterceptor(appMode string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		res, err := handler(ctx, req)
		if err != nil {
			return nil, gRPCError(err, appMode, req)
		}
		return res, nil
	}
}

// Todo ...
// StreamServerInterceptor returns a new streaming server interceptor that wraps outcome error.
//
// The stage at which invalid messages will be rejected with `InvalidArgument` varies based on the
// type of the RPC. For `ServerStream` (1:m) requests, it will happen before reaching any userspace
// handlers. For `ClientStream` (n:1) or `BidiStream` (n:m) RPCs, the messages will be rejected on
// calls to `stream.Recv()`.
//func StreamServerInterceptor() grpc.StreamServerInterceptor {
//	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
//		wrapper := &recvWrapper{stream}
//		return handler(srv, wrapper)
//	}
//}

// gRPCError converts original error to GRPC error which will then be converted to HTTP error by grpc-gateway.
func gRPCError(err error, appMode string, req interface{}) error {
	if errors.Is(err, context.Canceled) {
		return status.New(codes.Canceled, err.Error()).Err()
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return status.New(codes.DeadlineExceeded, err.Error()).Err()
	}

	var serr interface {
		error
		GRPCStatus() *status.Status
	}

	var stt *status.Status
	if errors.As(err, &serr) {
		stt = status.Convert(serr)
	} else {
		stt = status.Convert(err)
	}

	var derr interface {
		error
		Details() []proto.Message
	}

	if errors.As(err, &derr) {
		stt, _ = stt.WithDetails(derr.Details()...)
	}

	logger.WithFields(logger.Fields{"error": err}).Error("getting error...")

	if appMode == ProductionMode && (stt.Code() == codes.Unknown || stt.Code() == codes.Internal) {
		return &internalServerError{Err: err, Req: req}
	}

	return stt.Err()
}

// internalServerError wraps unknown/server errors and hide its detail from clients.
type internalServerError struct {
	Err error
	Req interface{}
}

func (e internalServerError) Error() string {
	return "Something went wrong in our side"
}
