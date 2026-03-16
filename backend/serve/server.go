package serve

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/internal/analytics"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/tracking"
	"github.com/yumikokawaii/sherry-archive/internal/handler"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
	"github.com/yumikokawaii/sherry-archive/internal/service"
	"github.com/yumikokawaii/sherry-archive/pkg/queue"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
	"github.com/yumikokawaii/sherry-archive/pkg/token"
	"github.com/yumikokawaii/sherry-archive/pkg/urlcache"
)

func Server(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Parse duration strings
	accessExpiry, err := time.ParseDuration(cfg.JWT.AccessTokenExpiry)
	if err != nil {
		log.Fatalf("invalid jwt.access_token_expiry %q: %v", cfg.JWT.AccessTokenExpiry, err)
	}
	refreshExpiry, err := time.ParseDuration(cfg.JWT.RefreshTokenExpiry)
	if err != nil {
		log.Fatalf("invalid jwt.refresh_token_expiry %q: %v", cfg.JWT.RefreshTokenExpiry, err)
	}
	presignExpiry, err := time.ParseDuration(cfg.S3.PresignExpiry)
	if err != nil {
		log.Fatalf("invalid s3.presign_expiry %q: %v", cfg.S3.PresignExpiry, err)
	}
	decayInterval, err := time.ParseDuration(cfg.Analytics.DecayInterval)
	if err != nil {
		log.Fatalf("invalid analytics.decay_interval %q: %v", cfg.Analytics.DecayInterval, err)
	}

	// Database
	db, err := postgres.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("db connect: %v", err)
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
		log.Fatalf("s3: %v", err)
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
		log.Fatalf("sqs: %v", err)
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
			log.Fatalf("cloudfront signer: %v", err)
		}
		signer = cfSigner
	}
	urlCache := urlcache.New(signer, rdb, presignExpiry)

	// Services
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, deviceMappingRepo, tokenMgr)
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
	}

	r := handler.SetupRouter(handlers, tokenMgr)

	// Background context cancelled on shutdown
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	// Analytics — real-time trending + suggestions via Redis
	stopTags := make(map[string]struct{})
	if cfg.Analytics.StopTags != "" {
		for _, t := range strings.Split(cfg.Analytics.StopTags, ",") {
			if t = strings.TrimSpace(t); t != "" {
				stopTags[t] = struct{}{}
			}
		}
	}
	analyticsStore := analytics.NewStore(rdb, db, cfg.Analytics.ContributionCap, decayInterval, stopTags)
	analytics.NewHandler(analyticsStore, urlCache).Mount(r)
	go analyticsStore.StartDecay(bgCtx)

	// Tracking — mounted independently; enriched by analytics store
	trackingStore := tracking.NewPostgresStore(db)
	tracking.NewHandler(trackingStore, tokenMgr, analyticsStore).Mount(r)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("server starting on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
