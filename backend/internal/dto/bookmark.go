package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

// --- Requests ---

type UpsertBookmarkRequest struct {
	ChapterID      uuid.UUID `json:"chapter_id"       binding:"required"`
	LastPageNumber int       `json:"last_page_number" binding:"required,min=1"`
}

// --- Responses ---

type BookmarkResponse struct {
	ID             uuid.UUID `json:"id"`
	MangaID        uuid.UUID `json:"manga_id"`
	ChapterID      uuid.UUID `json:"chapter_id"`
	LastPageNumber int       `json:"last_page_number"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func NewBookmarkResponse(b *model.Bookmark) BookmarkResponse {
	return BookmarkResponse{
		ID:             b.ID,
		MangaID:        b.MangaID,
		ChapterID:      b.ChapterID,
		LastPageNumber: b.LastPageNumber,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
	}
}

func NewBookmarkResponseList(bs []*model.Bookmark) []BookmarkResponse {
	out := make([]BookmarkResponse, len(bs))
	for i, b := range bs {
		out[i] = NewBookmarkResponse(b)
	}
	return out
}
