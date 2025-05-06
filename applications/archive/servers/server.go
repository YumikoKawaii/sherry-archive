package servers

import (
	"context"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sherry.archive.com/applications/archive/adapters/multimedia"
	"sherry.archive.com/applications/archive/config"
	"sherry.archive.com/applications/archive/pkg/repository"
	"sherry.archive.com/applications/archive/servers/apis"
	"sherry.archive.com/applications/archive/servers/file_processors/extractor"
	"sherry.archive.com/applications/archive/services"
	"sherry.archive.com/shared/database"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/middleware/grpc_error"
	"sherry.archive.com/shared/middleware/grpc_recovery"
	"sherry.archive.com/shared/topics"
)

func Serve(cfg *config.Application) {
	prometheus := grpc_prometheus.NewServerMetrics()
	zapSugaredLogger := logger.GetDelegate().(*zap.SugaredLogger)
	zapLogger := zapSugaredLogger.Desugar()
	mysqlGorm := database.NewMysqlGormDatabase(cfg.MysqlConfig.DSN())
	grpc_zap.ReplaceGrpcLoggerV2(zapLogger)

	querier := repository.NewQuerier(mysqlGorm)
	multimediaStorage := multimedia.NewCloudinaryClient(cfg.CloudinaryConfig)
	publisher := topics.NewKafkaSyncPublisher(cfg.KafkaConfig)
	service := apis.NewService(querier, multimediaStorage, publisher)

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

func Extract(cfg *config.Application) {
	consumer := topics.NewKafkaConsumer(cfg.KafkaConfig)
	publisher := topics.NewKafkaSyncPublisher(cfg.KafkaConfig)
	ext := extractor.NewExtractor(consumer, publisher)
	ext.Extract(context.Background())
}
