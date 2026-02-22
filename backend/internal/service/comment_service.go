package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/pkg/pagination"
)

type CommentService struct {
	commentRepo repository.CommentRepository
	mangaRepo   repository.MangaRepository
	chapterRepo repository.ChapterRepository
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	mangaRepo repository.MangaRepository,
	chapterRepo repository.ChapterRepository,
) *CommentService {
	return &CommentService{commentRepo: commentRepo, mangaRepo: mangaRepo, chapterRepo: chapterRepo}
}

func (s *CommentService) CreateMangaComment(ctx context.Context, userID, mangaID uuid.UUID, content string) (*model.CommentWithAuthor, error) {
	if _, err := s.mangaRepo.GetByID(ctx, mangaID); err != nil {
		return nil, err
	}

	c := &model.Comment{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		MangaID:   mangaID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return s.commentRepo.GetByID(ctx, c.ID)
}

func (s *CommentService) CreateChapterComment(ctx context.Context, userID, mangaID, chapterID uuid.UUID, content string) (*model.CommentWithAuthor, error) {
	ch, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	if ch.MangaID != mangaID {
		return nil, apperror.ErrNotFound
	}

	c := &model.Comment{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		MangaID:   mangaID,
		ChapterID: &chapterID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	return s.commentRepo.GetByID(ctx, c.ID)
}

func (s *CommentService) Update(ctx context.Context, userID, commentID uuid.UUID, content string) (*model.CommentWithAuthor, error) {
	c, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}
	if c.UserID != userID {
		return nil, apperror.ErrForbidden
	}

	c.Content = content
	c.Edited = true
	c.UpdatedAt = time.Now()
	if err := s.commentRepo.Update(ctx, &c.Comment); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CommentService) Delete(ctx context.Context, requesterID, commentID uuid.UUID) error {
	c, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}

	// Comment owner can always delete; manga owner can also delete
	if c.UserID == requesterID {
		return s.commentRepo.Delete(ctx, commentID)
	}
	manga, err := s.mangaRepo.GetByID(ctx, c.MangaID)
	if err != nil {
		return err
	}
	if manga.OwnerID != requesterID {
		return apperror.ErrForbidden
	}
	return s.commentRepo.Delete(ctx, commentID)
}

func (s *CommentService) ListByManga(ctx context.Context, mangaID uuid.UUID, p pagination.Params) ([]*model.CommentWithAuthor, int, error) {
	return s.commentRepo.ListByManga(ctx, mangaID, p)
}

func (s *CommentService) ListByChapter(ctx context.Context, mangaID, chapterID uuid.UUID, p pagination.Params) ([]*model.CommentWithAuthor, int, error) {
	ch, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return nil, 0, err
	}
	if ch.MangaID != mangaID {
		return nil, 0, apperror.ErrNotFound
	}
	return s.commentRepo.ListByChapter(ctx, chapterID, p)
}
