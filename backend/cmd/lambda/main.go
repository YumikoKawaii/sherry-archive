package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
	"github.com/yumikokawaii/sherry-archive/internal/service"
	"github.com/yumikokawaii/sherry-archive/pkg/logger"
	"github.com/yumikokawaii/sherry-archive/pkg/queue"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
)

// Initialized once on cold start, reused across warm invocations.
var (
	pageSvc        *service.PageService
	uploadTaskRepo repository.UploadTaskRepository
	storageClient  *storage.Client
)

func init() {
	logger.Init()

	zap.L().Info("init: loading config")
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("init: config", zap.Error(err))
	}

	zap.L().Info("init: connecting to database")
	db, err := postgres.Connect(cfg.DB.DSN())
	if err != nil {
		zap.L().Fatal("init: db", zap.Error(err))
	}

	zap.L().Info("init: connecting to S3", zap.String("bucket", cfg.S3.Bucket), zap.String("region", cfg.S3.Region))
	sc, err := storage.NewClient(context.Background(), cfg.S3.Region, cfg.S3.Bucket, cfg.S3.Endpoint, time.Hour)
	if err != nil {
		zap.L().Fatal("init: s3", zap.Error(err))
	}

	pageRepo := postgres.NewPageRepo(db)
	chapterRepo := postgres.NewChapterRepo(db)
	mangaRepo := postgres.NewMangaRepo(db)

	// urlCache is nil — Lambda only calls UploadZip/UploadOneshotZip which don't use it.
	pageSvc = service.NewPageService(pageRepo, chapterRepo, mangaRepo, sc, nil)
	uploadTaskRepo = postgres.NewUploadTaskRepo(db)
	storageClient = sc
	zap.L().Info("init: ready")
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		if err := processRecord(ctx, record); err != nil {
			// Returning an error causes SQS to retry (or route to DLQ after max attempts).
			return err
		}
	}
	return nil
}

func processRecord(ctx context.Context, record events.SQSMessage) error {
	var msg queue.UploadMessage
	if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
		return fmt.Errorf("unmarshal message: %w", err)
	}

	zap.L().Info("task received", zap.String("task_id", msg.TaskID.String()), zap.String("type", msg.Type), zap.String("manga_id", msg.MangaID.String()))

	// Atomically claim the task (pending → processing).
	// Returns false if another Lambda already claimed it or it's already done — skip silently.
	claimed, err := uploadTaskRepo.ClaimProcessing(ctx, msg.TaskID)
	if err != nil {
		return fmt.Errorf("claim task: %w", err)
	}
	if !claimed {
		zap.L().Info("task already claimed or done, skipping", zap.String("task_id", msg.TaskID.String()))
		return nil
	}

	zap.L().Info("task claimed, processing", zap.String("task_id", msg.TaskID.String()))
	if err := process(ctx, msg); err != nil {
		zap.L().Error("task failed", zap.String("task_id", msg.TaskID.String()), zap.Error(err))
		_ = uploadTaskRepo.UpdateStatus(ctx, msg.TaskID, model.UploadTaskStatusFailed, err.Error())
		return err
	}
	zap.L().Info("task done", zap.String("task_id", msg.TaskID.String()))
	return nil
}

func process(ctx context.Context, msg queue.UploadMessage) error {
	zap.L().Info("downloading staging zip", zap.String("task_id", msg.TaskID.String()), zap.String("s3_key", msg.S3Key))
	body, err := storageClient.GetObject(ctx, msg.S3Key)
	if err != nil {
		return fmt.Errorf("get staging zip: %w", err)
	}
	defer body.Close()

	tmp, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	size, err := io.Copy(tmp, body)
	if err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	zap.L().Info("zip downloaded", zap.String("task_id", msg.TaskID.String()), zap.Int64("size_bytes", size))

	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	switch model.UploadTaskType(msg.Type) {
	case model.UploadTaskTypeZip:
		if msg.ChapterID == nil {
			return fmt.Errorf("chapter_id required for zip task")
		}
		zap.L().Info("processing zip", zap.String("task_id", msg.TaskID.String()), zap.String("chapter_id", msg.ChapterID.String()))
		if _, _, err := pageSvc.UploadZip(ctx, msg.OwnerID, msg.MangaID, *msg.ChapterID, tmp, size); err != nil {
			return err
		}
		_ = uploadTaskRepo.UpdateStatus(ctx, msg.TaskID, model.UploadTaskStatusDone, "")

	case model.UploadTaskTypeOneshotZip:
		zap.L().Info("processing oneshot zip", zap.String("task_id", msg.TaskID.String()), zap.String("manga_id", msg.MangaID.String()))
		result, err := pageSvc.UploadOneshotZip(ctx, msg.OwnerID, msg.MangaID, tmp, size)
		if err != nil {
			return err
		}
		zap.L().Info("oneshot chapter created", zap.String("task_id", msg.TaskID.String()), zap.String("chapter_id", result.Chapter.ID.String()))
		_ = uploadTaskRepo.SetChapterAndDone(ctx, msg.TaskID, result.Chapter.ID)

	default:
		return fmt.Errorf("unknown task type: %s", msg.Type)
	}

	zap.L().Info("deleting staging zip", zap.String("task_id", msg.TaskID.String()))
	_ = storageClient.DeleteObject(ctx, msg.S3Key)
	return nil
}
