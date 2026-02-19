package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	pkgslug "github.com/yumikokawaii/sherry-archive/pkg/slug"
	"github.com/yumikokawaii/sherry-archive/pkg/pagination"
)

type MangaService struct {
	mangaRepo repository.MangaRepository
}

func NewMangaService(mangaRepo repository.MangaRepository) *MangaService {
	return &MangaService{mangaRepo: mangaRepo}
}

type CreateMangaInput struct {
	OwnerID     uuid.UUID
	Title       string
	Description string
	Status      model.MangaStatus
	Type        model.MangaType
	Tags        []string
	Author      string
	Artist      string
	Category    string
}

func (s *MangaService) Create(ctx context.Context, in CreateMangaInput) (*model.Manga, error) {
	slug, err := s.uniqueSlug(ctx, in.Title)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	m := &model.Manga{
		ID:          uuid.Must(uuid.NewV7()),
		OwnerID:     in.OwnerID,
		Title:       in.Title,
		Slug:        slug,
		Description: in.Description,
		Status:      in.Status,
		Type:        in.Type,
		Tags:        pq.StringArray(in.Tags),
		Author:      in.Author,
		Artist:      in.Artist,
		Category:    in.Category,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.mangaRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

type UpdateMangaInput struct {
	Title       *string
	Description *string
	Status      *model.MangaStatus
	Type        *model.MangaType
	Tags        []string
	Author      *string
	Artist      *string
	Category    *string
}

func (s *MangaService) Update(ctx context.Context, requesterID, mangaID uuid.UUID, in UpdateMangaInput) (*model.Manga, error) {
	m, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	if m.OwnerID != requesterID {
		return nil, apperror.ErrForbidden
	}

	if in.Title != nil && *in.Title != m.Title {
		slug, err := s.uniqueSlug(ctx, *in.Title)
		if err != nil {
			return nil, err
		}
		m.Title = *in.Title
		m.Slug = slug
	}
	if in.Description != nil {
		m.Description = *in.Description
	}
	if in.Status != nil {
		m.Status = *in.Status
	}
	if in.Type != nil {
		m.Type = *in.Type
	}
	if in.Tags != nil {
		m.Tags = pq.StringArray(in.Tags)
	}
	if in.Author != nil {
		m.Author = *in.Author
	}
	if in.Artist != nil {
		m.Artist = *in.Artist
	}
	if in.Category != nil {
		m.Category = *in.Category
	}
	m.UpdatedAt = time.Now()

	if err := s.mangaRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MangaService) Delete(ctx context.Context, requesterID, mangaID uuid.UUID) error {
	m, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return err
	}
	if m.OwnerID != requesterID {
		return apperror.ErrForbidden
	}
	return s.mangaRepo.Delete(ctx, mangaID)
}

func (s *MangaService) GetByID(ctx context.Context, id uuid.UUID) (*model.Manga, error) {
	return s.mangaRepo.GetByID(ctx, id)
}

func (s *MangaService) List(ctx context.Context, filter repository.MangaFilter, p pagination.Params) ([]*model.Manga, int, error) {
	return s.mangaRepo.List(ctx, filter, p)
}

func (s *MangaService) ListByOwner(ctx context.Context, ownerID uuid.UUID, p pagination.Params) ([]*model.Manga, int, error) {
	return s.mangaRepo.ListByOwner(ctx, ownerID, p)
}

func (s *MangaService) UpdateCover(ctx context.Context, requesterID, mangaID uuid.UUID, coverKey string) (*model.Manga, error) {
	m, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	if m.OwnerID != requesterID {
		return nil, apperror.ErrForbidden
	}
	m.CoverKey = coverKey
	m.UpdatedAt = time.Now()
	if err := s.mangaRepo.Update(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *MangaService) uniqueSlug(ctx context.Context, title string) (string, error) {
	base := pkgslug.Make(title)
	slug := base
	for i := 1; ; i++ {
		exists, err := s.mangaRepo.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}
