package servers

import (
	"encoding/hex"
	"github.com/go-redis/redis/v8"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sherry.archive.com/applications/iam/config"
	"sherry.archive.com/applications/iam/pkg/jwt"
	"sherry.archive.com/applications/iam/pkg/repository"
	"sherry.archive.com/applications/iam/servers/apis"
	"sherry.archive.com/applications/iam/services"
	"sherry.archive.com/shared/database"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/middleware/grpc_error"
	"sherry.archive.com/shared/middleware/grpc_recovery"
)

func Serve(cfg *config.Application) {
	prometheus := grpc_prometheus.NewServerMetrics()
	zapSugaredLogger := logger.GetDelegate().(*zap.SugaredLogger)
	zapLogger := zapSugaredLogger.Desugar()
	mysqlGorm := database.NewMysqlGormDatabase(cfg.MysqlConfig.DSN())
	grpc_zap.ReplaceGrpcLoggerV2(zapLogger)
	hashKey, _ := hex.DecodeString(cfg.HashKey)
	serviceCfg := &apis.Config{
		HashKey: hashKey,
	}
	querier := repository.NewQuerier(mysqlGorm)
	jwtResolver := jwt.NewResolver(cfg.JwtKey, cfg.JwtTTLInSec)
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})
	service := apis.NewService(serviceCfg, querier, jwtResolver, redisClient)

	grpc_prometheus.EnableHandlingTimeHistogram()
	sv := services.NewServer(
		services.NewConfig(cfg.GRPCPort, cfg.HTTPPort),
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.ChainUnaryInterceptor(
			prometheus.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_zap.UnaryServerInterceptor(zapLogger),
			grpc_validator.UnaryServerInterceptor(),
			grpc_error.UnaryServerInterceptor(cfg.AppMode),
			grpc_recovery.UnaryServerInterceptor(),
		),
	)

	if err := sv.Register(service); err != nil {
		logger.Fatalf("error register server: %s", err.Error())
	}

	if err := sv.Serve(); err != nil {
		logger.Fatalf("failed to serve: %s", err.Error())
	}
}
