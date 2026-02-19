package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/pkg/pagination"
)

type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, u *model.User) error
}

type MangaFilter struct {
	Query  string
	Status string
	Tags   []string
	Sort   string // "newest" | "oldest" | "title"
}

type MangaRepository interface {
	Create(ctx context.Context, m *model.Manga) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Manga, error)
	GetBySlug(ctx context.Context, slug string) (*model.Manga, error)
	List(ctx context.Context, filter MangaFilter, p pagination.Params) ([]*model.Manga, int, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, p pagination.Params) ([]*model.Manga, int, error)
	Update(ctx context.Context, m *model.Manga) error
	Delete(ctx context.Context, id uuid.UUID) error
	SlugExists(ctx context.Context, slug string) (bool, error)
}

type ChapterRepository interface {
	Create(ctx context.Context, ch *model.Chapter) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Chapter, error)
	GetByMangaAndNumber(ctx context.Context, mangaID uuid.UUID, number float64) (*model.Chapter, error)
	ListByManga(ctx context.Context, mangaID uuid.UUID) ([]*model.Chapter, error)
	Update(ctx context.Context, ch *model.Chapter) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePageCount(ctx context.Context, id uuid.UUID, count int) error
}

type PageRepository interface {
	CreateBatch(ctx context.Context, pages []*model.Page) error
	GetByChapter(ctx context.Context, chapterID uuid.UUID) ([]*model.Page, error)
	GetByChapterAndNumber(ctx context.Context, chapterID uuid.UUID, number int) (*model.Page, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateNumbers(ctx context.Context, chapterID uuid.UUID, pageIDs []uuid.UUID) error
	CountByChapter(ctx context.Context, chapterID uuid.UUID) (int, error)
}

type BookmarkRepository interface {
	Upsert(ctx context.Context, b *model.Bookmark) error
	GetByUserAndManga(ctx context.Context, userID, mangaID uuid.UUID) (*model.Bookmark, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Bookmark, error)
	Delete(ctx context.Context, userID, mangaID uuid.UUID) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *model.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	DeleteByHash(ctx context.Context, hash string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
