package serve

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/handler"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
	"github.com/yumikokawaii/sherry-archive/internal/service"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
	"github.com/yumikokawaii/sherry-archive/pkg/token"
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
	presignExpiry, err := time.ParseDuration(cfg.MinIO.PresignExpiry)
	if err != nil {
		log.Fatalf("invalid minio.presign_expiry %q: %v", cfg.MinIO.PresignExpiry, err)
	}

	// Database
	db, err := postgres.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db.DB); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// MinIO
	storageClient, err := storage.NewClient(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKeyID,
		cfg.MinIO.SecretAccessKey,
		cfg.MinIO.Bucket,
		cfg.MinIO.UseSSL,
		presignExpiry,
	)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}
	if err := storageClient.EnsureBucket(context.Background()); err != nil {
		log.Fatalf("minio ensure bucket: %v", err)
	}

	// Token manager
	tokenMgr := token.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		accessExpiry,
		refreshExpiry,
	)

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	mangaRepo := postgres.NewMangaRepo(db)
	chapterRepo := postgres.NewChapterRepo(db)
	pageRepo := postgres.NewPageRepo(db)
	bookmarkRepo := postgres.NewBookmarkRepo(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, tokenMgr)
	userSvc := service.NewUserService(userRepo)
	mangaSvc := service.NewMangaService(mangaRepo)
	chapterSvc := service.NewChapterService(chapterRepo, mangaRepo)
	pageSvc := service.NewPageService(pageRepo, chapterRepo, mangaRepo, storageClient)
	bookmarkSvc := service.NewBookmarkService(bookmarkRepo)

	// Handlers
	handlers := handler.Handlers{
		Auth:     handler.NewAuthHandler(authSvc),
		Manga:    handler.NewMangaHandler(mangaSvc, storageClient),
		Chapter:  handler.NewChapterHandler(chapterSvc, pageSvc),
		Page:     handler.NewPageHandler(pageSvc),
		Bookmark: handler.NewBookmarkHandler(bookmarkSvc),
		User:     handler.NewUserHandler(userSvc, storageClient),
	}

	r := handler.SetupRouter(handlers, tokenMgr)

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

func runMigrations(db *sql.DB) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
