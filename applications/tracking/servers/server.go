package servers

import (
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sherry.archive.com/applications/tracking/config"
	"sherry.archive.com/applications/tracking/servers/apis"
	"sherry.archive.com/applications/tracking/services"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/middleware/grpc_error"
	"sherry.archive.com/shared/middleware/grpc_recovery"
)

func Serve(cfg *config.Application) {
	prometheus := grpc_prometheus.NewServerMetrics()
	zapSugaredLogger := logger.GetDelegate().(*zap.SugaredLogger)
	zapLogger := zapSugaredLogger.Desugar()
	//mysqlGorm := database.NewMysqlGormDatabase(cfg.MysqlConfig.DSN())
	grpc_zap.ReplaceGrpcLoggerV2(zapLogger)
	
	service := apis.NewService()

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
