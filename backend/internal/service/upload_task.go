package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/pkg/queue"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
)

type UploadTaskService struct {
	uploadTaskRepo repository.UploadTaskRepository
	mangaRepo      repository.MangaRepository
	chapterRepo    repository.ChapterRepository
	storage        *storage.Client
	queue          *queue.Client
}

func NewUploadTaskService(
	uploadTaskRepo repository.UploadTaskRepository,
	mangaRepo repository.MangaRepository,
	chapterRepo repository.ChapterRepository,
	storage *storage.Client,
	queue *queue.Client,
) *UploadTaskService {
	return &UploadTaskService{
		uploadTaskRepo: uploadTaskRepo,
		mangaRepo:      mangaRepo,
		chapterRepo:    chapterRepo,
		storage:        storage,
		queue:          queue,
	}
}

// EnqueueZipUpload validates the zip, stages it to S3, and sends an SQS message.
// Returns 202 immediately; Lambda processes asynchronously.
func (s *UploadTaskService) EnqueueZipUpload(ctx context.Context, requesterID, mangaID, chapterID uuid.UUID, r io.Reader) (*model.UploadTask, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := validateZip(data); err != nil {
		return nil, err
	}

	manga, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	if manga.OwnerID != requesterID {
		return nil, apperror.ErrForbidden
	}
	ch, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	if ch.MangaID != mangaID {
		return nil, apperror.ErrNotFound
	}

	return s.enqueue(ctx, model.UploadTaskTypeZip, requesterID, mangaID, &chapterID, data)
}

// EnqueueOneshotZipUpload validates the zip, stages it to S3, and sends an SQS message.
// The chapter is created by Lambda; chapter_id on the task starts as NULL.
func (s *UploadTaskService) EnqueueOneshotZipUpload(ctx context.Context, requesterID, mangaID uuid.UUID, r io.Reader) (*model.UploadTask, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := validateZip(data); err != nil {
		return nil, err
	}

	manga, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	if manga.OwnerID != requesterID {
		return nil, apperror.ErrForbidden
	}
	if manga.Type != model.TypeOneshot {
		return nil, apperror.ErrBadRequest
	}

	return s.enqueue(ctx, model.UploadTaskTypeOneshotZip, requesterID, mangaID, nil, data)
}

func (s *UploadTaskService) GetTask(ctx context.Context, id uuid.UUID) (*model.UploadTask, error) {
	return s.uploadTaskRepo.GetByID(ctx, id)
}

func (s *UploadTaskService) enqueue(ctx context.Context, taskType model.UploadTaskType, ownerID, mangaID uuid.UUID, chapterID *uuid.UUID, data []byte) (*model.UploadTask, error) {
	now := time.Now()
	task := &model.UploadTask{
		ID:        uuid.Must(uuid.NewV7()),
		Type:      taskType,
		Status:    model.UploadTaskStatusPending,
		OwnerID:   ownerID,
		MangaID:   mangaID,
		S3Key:     fmt.Sprintf("uploads/%s.zip", uuid.Must(uuid.NewV7())),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if chapterID != nil {
		task.ChapterID = uuid.NullUUID{UUID: *chapterID, Valid: true}
	}

	if err := s.uploadTaskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// bytes.NewReader is io.ReadSeeker — PutObject skips the extra io.ReadAll
	if err := s.storage.PutObject(ctx, task.S3Key, "application/zip", bytes.NewReader(data), int64(len(data))); err != nil {
		return nil, err
	}

	msg := queue.UploadMessage{
		TaskID:    task.ID,
		Type:      string(task.Type),
		S3Key:     task.S3Key,
		MangaID:   mangaID,
		OwnerID:   ownerID,
		ChapterID: chapterID,
	}
	if err := s.queue.Enqueue(ctx, msg); err != nil {
		return nil, err
	}

	return task, nil
}

// validateZip checks the bytes are a parseable zip with at least one image entry.
func validateZip(data []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return apperror.ErrBadRequest
	}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if _, ok := allowedExtensions[ext]; ok {
			return nil
		}
	}
	return apperror.ErrBadRequest
}
