package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

// --- Requests ---

type CreateChapterRequest struct {
	Number float64 `json:"number" binding:"required"`
	Title  string  `json:"title"`
}

type UpdateChapterRequest struct {
	Number *float64 `json:"number"`
	Title  *string  `json:"title"`
}

// --- Responses ---

type ChapterResponse struct {
	ID        uuid.UUID `json:"id"`
	MangaID   uuid.UUID `json:"manga_id"`
	Number    float64   `json:"number"`
	Title     string    `json:"title"`
	PageCount int       `json:"page_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PageItemResponse struct {
	ID     uuid.UUID `json:"id"`
	Number int       `json:"number"`
	URL    string    `json:"url"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
}

type ChapterWithPagesResponse struct {
	Chapter ChapterResponse    `json:"chapter"`
	Pages   []PageItemResponse `json:"pages"`
}

func NewChapterResponse(ch *model.Chapter) ChapterResponse {
	return ChapterResponse{
		ID:        ch.ID,
		MangaID:   ch.MangaID,
		Number:    ch.Number,
		Title:     ch.Title,
		PageCount: ch.PageCount,
		CreatedAt: ch.CreatedAt,
		UpdatedAt: ch.UpdatedAt,
	}
}

func NewChapterResponseList(chs []*model.Chapter) []ChapterResponse {
	out := make([]ChapterResponse, len(chs))
	for i, ch := range chs {
		out[i] = NewChapterResponse(ch)
	}
	return out
}
