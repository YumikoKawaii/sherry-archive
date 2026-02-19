package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
)

type BookmarkService struct {
	bookmarkRepo repository.BookmarkRepository
}

func NewBookmarkService(bookmarkRepo repository.BookmarkRepository) *BookmarkService {
	return &BookmarkService{bookmarkRepo: bookmarkRepo}
}

type UpsertBookmarkInput struct {
	ChapterID      uuid.UUID
	LastPageNumber int
}

func (s *BookmarkService) Upsert(ctx context.Context, userID, mangaID uuid.UUID, in UpsertBookmarkInput) (*model.Bookmark, error) {
	now := time.Now()
	b := &model.Bookmark{
		ID:             uuid.Must(uuid.NewV7()),
		UserID:         userID,
		MangaID:        mangaID,
		ChapterID:      in.ChapterID,
		LastPageNumber: in.LastPageNumber,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.bookmarkRepo.Upsert(ctx, b); err != nil {
		return nil, err
	}
	// Re-fetch to get the actual (potentially existing) record after upsert
	return s.bookmarkRepo.GetByUserAndManga(ctx, userID, mangaID)
}

func (s *BookmarkService) Get(ctx context.Context, userID, mangaID uuid.UUID) (*model.Bookmark, error) {
	return s.bookmarkRepo.GetByUserAndManga(ctx, userID, mangaID)
}

func (s *BookmarkService) List(ctx context.Context, userID uuid.UUID) ([]*model.Bookmark, error) {
	return s.bookmarkRepo.ListByUser(ctx, userID)
}

func (s *BookmarkService) Delete(ctx context.Context, userID, mangaID uuid.UUID) error {
	return s.bookmarkRepo.Delete(ctx, userID, mangaID)
}
