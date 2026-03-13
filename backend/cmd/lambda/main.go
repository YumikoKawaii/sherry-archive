package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
	"github.com/yumikokawaii/sherry-archive/internal/service"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := postgres.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	// presignExpiry is unused in Lambda (we never generate presigned URLs here),
	// but NewClient requires it — pass any non-zero value.
	sc, err := storage.NewClient(context.Background(), cfg.S3.Region, cfg.S3.Bucket, cfg.S3.Endpoint, time.Hour)
	if err != nil {
		log.Fatalf("s3: %v", err)
	}

	pageRepo := postgres.NewPageRepo(db)
	chapterRepo := postgres.NewChapterRepo(db)
	mangaRepo := postgres.NewMangaRepo(db)

	// urlCache is nil — Lambda only calls UploadZip/UploadOneshotZip which don't use it.
	pageSvc = service.NewPageService(pageRepo, chapterRepo, mangaRepo, sc, nil)
	uploadTaskRepo = postgres.NewUploadTaskRepo(db)
	storageClient = sc
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

	_ = uploadTaskRepo.UpdateStatus(ctx, msg.TaskID, model.UploadTaskStatusProcessing, "")

	if err := process(ctx, msg); err != nil {
		_ = uploadTaskRepo.UpdateStatus(ctx, msg.TaskID, model.UploadTaskStatusFailed, err.Error())
		return err
	}
	return nil
}

func process(ctx context.Context, msg queue.UploadMessage) error {
	// Download staging zip to a temp file so we have io.ReaderAt for zip.NewReader.
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
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return err
	}

	switch model.UploadTaskType(msg.Type) {
	case model.UploadTaskTypeZip:
		if msg.ChapterID == nil {
			return fmt.Errorf("chapter_id required for zip task")
		}
		if _, _, err := pageSvc.UploadZip(ctx, msg.OwnerID, msg.MangaID, *msg.ChapterID, tmp, size); err != nil {
			return err
		}
		_ = uploadTaskRepo.UpdateStatus(ctx, msg.TaskID, model.UploadTaskStatusDone, "")

	case model.UploadTaskTypeOneshotZip:
		result, err := pageSvc.UploadOneshotZip(ctx, msg.OwnerID, msg.MangaID, tmp, size)
		if err != nil {
			return err
		}
		_ = uploadTaskRepo.SetChapterAndDone(ctx, msg.TaskID, result.Chapter.ID)

	default:
		return fmt.Errorf("unknown task type: %s", msg.Type)
	}

	// Delete staging zip — best effort, don't fail the task if this errors.
	_ = storageClient.DeleteObject(ctx, msg.S3Key)
	return nil
}
