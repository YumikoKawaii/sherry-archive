package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
)

type ChapterService struct {
	chapterRepo repository.ChapterRepository
	mangaRepo   repository.MangaRepository
}

func NewChapterService(
	chapterRepo repository.ChapterRepository,
	mangaRepo repository.MangaRepository,
) *ChapterService {
	return &ChapterService{chapterRepo: chapterRepo, mangaRepo: mangaRepo}
}

type CreateChapterInput struct {
	MangaID uuid.UUID
	Number  *float64
	Title   string
}

func (s *ChapterService) Create(ctx context.Context, requesterID uuid.UUID, in CreateChapterInput) (*model.Chapter, error) {
	manga, err := s.mangaRepo.GetByID(ctx, in.MangaID)
	if err != nil {
		return nil, err
	}
	if manga.OwnerID != requesterID {
		return nil, apperror.ErrForbidden
	}

	var number float64
	title := in.Title

	if manga.Type == model.TypeOneshot {
		// Enforce max 1 chapter for oneshot
		existing, err := s.chapterRepo.ListByManga(ctx, in.MangaID)
		if err != nil {
			return nil, err
		}
		if len(existing) > 0 {
			return nil, apperror.ErrConflict
		}
		number = 0
		if title == "" {
			title = "Oneshot"
		}
	} else {
		if in.Number == nil {
			return nil, apperror.ErrBadRequest
		}
		number = *in.Number
		// Check duplicate chapter number
		if existing, err := s.chapterRepo.GetByMangaAndNumber(ctx, in.MangaID, number); err == nil && existing != nil {
			return nil, apperror.ErrConflict
		}
	}

	now := time.Now()
	ch := &model.Chapter{
		ID:        uuid.Must(uuid.NewV7()),
		MangaID:   in.MangaID,
		Number:    number,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.chapterRepo.Create(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

type UpdateChapterInput struct {
	Number *float64
	Title  *string
}

func (s *ChapterService) Update(ctx context.Context, requesterID, chapterID uuid.UUID, in UpdateChapterInput) (*model.Chapter, error) {
	ch, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	manga, err := s.mangaRepo.GetByID(ctx, ch.MangaID)
	if err != nil {
		return nil, err
	}
	if manga.OwnerID != requesterID {
		return nil, apperror.ErrForbidden
	}

	if in.Number != nil {
		ch.Number = *in.Number
	}
	if in.Title != nil {
		ch.Title = *in.Title
	}
	ch.UpdatedAt = time.Now()

	if err := s.chapterRepo.Update(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *ChapterService) Delete(ctx context.Context, requesterID, chapterID uuid.UUID) error {
	ch, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return err
	}
	manga, err := s.mangaRepo.GetByID(ctx, ch.MangaID)
	if err != nil {
		return err
	}
	if manga.OwnerID != requesterID {
		return apperror.ErrForbidden
	}
	return s.chapterRepo.Delete(ctx, chapterID)
}

func (s *ChapterService) GetByID(ctx context.Context, id uuid.UUID) (*model.Chapter, error) {
	return s.chapterRepo.GetByID(ctx, id)
}

func (s *ChapterService) ListByManga(ctx context.Context, mangaID uuid.UUID) ([]*model.Chapter, error) {
	// Verify manga exists
	if _, err := s.mangaRepo.GetByID(ctx, mangaID); err != nil {
		return nil, err
	}
	return s.chapterRepo.ListByManga(ctx, mangaID)
}
