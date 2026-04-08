package serve

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/internal/analytics"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/handler"
	"github.com/yumikokawaii/sherry-archive/internal/metrics"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
	"github.com/yumikokawaii/sherry-archive/internal/service"
	"github.com/yumikokawaii/sherry-archive/internal/tracking"
	"github.com/yumikokawaii/sherry-archive/internal/tracing"
	"github.com/yumikokawaii/sherry-archive/pkg/queue"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
	"github.com/yumikokawaii/sherry-archive/pkg/token"
	"github.com/yumikokawaii/sherry-archive/pkg/urlcache"
	"github.com/yumikokawaii/sherry-archive/pkg/xrayhook"
	"go.uber.org/zap"
)

func Server(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("config", zap.Error(err))
	}

	// Tracing — initialise X-Ray before connecting to DB so the instrumented driver is ready.
	if err := tracing.Init(cfg.Tracing); err != nil {
		zap.L().Fatal("tracing", zap.Error(err))
	}

	// Parse duration strings
	accessExpiry, err := time.ParseDuration(cfg.JWT.AccessTokenExpiry)
	if err != nil {
		zap.L().Fatal("invalid jwt.access_token_expiry", zap.String("value", cfg.JWT.AccessTokenExpiry), zap.Error(err))
	}
	refreshExpiry, err := time.ParseDuration(cfg.JWT.RefreshTokenExpiry)
	if err != nil {
		zap.L().Fatal("invalid jwt.refresh_token_expiry", zap.String("value", cfg.JWT.RefreshTokenExpiry), zap.Error(err))
	}
	presignExpiry, err := time.ParseDuration(cfg.S3.PresignExpiry)
	if err != nil {
		zap.L().Fatal("invalid s3.presign_expiry", zap.String("value", cfg.S3.PresignExpiry), zap.Error(err))
	}
	decayInterval, err := time.ParseDuration(cfg.Analytics.DecayInterval)
	if err != nil {
		zap.L().Fatal("invalid analytics.decay_interval", zap.String("value", cfg.Analytics.DecayInterval), zap.Error(err))
	}

	// Database — use X-Ray instrumented connection when tracing is enabled.
	var db *sqlx.DB
	if cfg.Tracing.Enabled {
		db, err = postgres.ConnectXRay(cfg.DB.DSN())
	} else {
		db, err = postgres.Connect(cfg.DB.DSN())
	}
	if err != nil {
		zap.L().Fatal("db connect", zap.Error(err))
	}
	defer db.Close()

	// S3
	storageClient, err := storage.NewClient(
		context.Background(),
		cfg.S3.Region,
		cfg.S3.Bucket,
		cfg.S3.Endpoint,
		presignExpiry,
	)
	if err != nil {
		zap.L().Fatal("s3", zap.Error(err))
	}

	// Redis
	redisOpts := &redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if cfg.Redis.TLS {
		redisOpts.TLSConfig = &tls.Config{}
	}
	rdb := redis.NewClient(redisOpts)
	defer rdb.Close()
	if cfg.Tracing.Enabled {
		rdb.AddHook(xrayhook.New())
	}

	// Token manager
	tokenMgr := token.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		accessExpiry,
		refreshExpiry,
	)

	// SQS
	sqsClient, err := queue.NewClient(context.Background(), cfg.S3.Region, cfg.SQS.QueueURL)
	if err != nil {
		zap.L().Fatal("sqs", zap.Error(err))
	}

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	mangaRepo := postgres.NewMangaRepo(db)
	chapterRepo := postgres.NewChapterRepo(db)
	pageRepo := postgres.NewPageRepo(db)
	bookmarkRepo := postgres.NewBookmarkRepo(db)
	commentRepo := postgres.NewCommentRepo(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepo(db)
	uploadTaskRepo := postgres.NewUploadTaskRepo(db)
	deviceMappingRepo := postgres.NewDeviceUserMappingRepo(db)
	seenMangaRepo := postgres.NewSeenMangaRepo(db)
	userInterestRepo := postgres.NewUserInterestRepo(db)

	// URL signer — CloudFront when configured, S3 presign otherwise
	var signer urlcache.Signer = storageClient
	if cfg.CloudFront.Domain != "" {
		cfSigner, err := storage.NewCloudFrontSigner(
			cfg.CloudFront.Domain,
			cfg.CloudFront.KeyPairID,
			cfg.CloudFront.PrivateKey,
			presignExpiry,
		)
		if err != nil {
			zap.L().Fatal("cloudfront signer", zap.Error(err))
		}
		signer = cfSigner
	}
	urlCache := urlcache.New(signer, rdb, presignExpiry)

	// Analytics — real-time trending + suggestions via Redis
	stopTags := make(map[string]struct{})
	if cfg.Analytics.StopTags != "" {
		for _, t := range strings.Split(cfg.Analytics.StopTags, ",") {
			if t = strings.TrimSpace(t); t != "" {
				stopTags[t] = struct{}{}
			}
		}
	}
	analyticsStore := analytics.NewStore(rdb, db, seenMangaRepo, cfg.Analytics.ContributionCap, decayInterval, stopTags)

	// Services
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, deviceMappingRepo, seenMangaRepo, userInterestRepo, analyticsStore, tokenMgr)
	userSvc := service.NewUserService(userRepo)
	mangaSvc := service.NewMangaService(mangaRepo)
	chapterSvc := service.NewChapterService(chapterRepo, mangaRepo)
	pageSvc := service.NewPageService(pageRepo, chapterRepo, mangaRepo, storageClient, urlCache)
	bookmarkSvc := service.NewBookmarkService(bookmarkRepo)
	commentSvc := service.NewCommentService(commentRepo, mangaRepo, chapterRepo)
	uploadTaskSvc := service.NewUploadTaskService(uploadTaskRepo, mangaRepo, chapterRepo, storageClient, sqsClient)

	// Handlers
	handlers := handler.Handlers{
		Auth:       handler.NewAuthHandler(authSvc),
		Manga:      handler.NewMangaHandler(mangaSvc, storageClient, urlCache),
		Chapter:    handler.NewChapterHandler(chapterSvc, pageSvc),
		Page:       handler.NewPageHandler(pageSvc, uploadTaskSvc),
		Bookmark:   handler.NewBookmarkHandler(bookmarkSvc),
		User:       handler.NewUserHandler(userSvc, storageClient),
		Comment:    handler.NewCommentHandler(commentSvc),
		UploadTask: handler.NewUploadTaskHandler(uploadTaskSvc),
		Sitemap:    handler.NewSitemapHandler(mangaRepo, chapterRepo),
	}

	r := handler.SetupRouter(handlers, tokenMgr)

	// Background context cancelled on shutdown
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	analytics.NewHandler(analyticsStore, urlCache).Mount(r)
	go analyticsStore.StartDecay(bgCtx)

	// Tracking — mounted independently; enriched by analytics store
	trackingStore := tracking.NewPostgresStore(db)
	tracking.NewHandler(trackingStore, tokenMgr, analyticsStore).Mount(r)

	// Metrics — push to CloudWatch every 60s; gated by metrics.enabled config
	if cfg.Metrics.Enabled {
		if err := metrics.Init(bgCtx, cfg.S3.Region, "SherryArchive", db.DB); err != nil {
			zap.L().Warn("metrics: cloudwatch unavailable, disabled", zap.Error(err))
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		zap.L().Info("server starting", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("forced shutdown", zap.Error(err))
	}
	zap.L().Info("server stopped")
}
