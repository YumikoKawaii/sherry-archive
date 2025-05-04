package grpc_recovery

import (
	"runtime/debug"
	"sherry.archive.com/shared/logger"

	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	internalServerErr = status.Error(codes.Internal, "Something went wrong in our side.")
	recoveryOpt       = recovery.WithRecoveryHandler(func(err interface{}) error {
		logger.WithFields(logger.Fields{"panic error": err, "stacktrace": string(debug.Stack())}).Error("unexpected error...")
		return internalServerErr
	})
)

// UnaryServerInterceptor ...
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(recoveryOpt)
}

// StreamServerInterceptor ...
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return recovery.StreamServerInterceptor(recoveryOpt)
}
