package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

// --- Requests ---

type CreateMangaRequest struct {
	Title       string            `json:"title"       binding:"required,min=1"`
	Description string            `json:"description"`
	Status      model.MangaStatus `json:"status"`
	Type        model.MangaType   `json:"type"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	Artist      string            `json:"artist"`
	Category    string            `json:"category"`
}

type UpdateMangaRequest struct {
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	Status      *model.MangaStatus `json:"status"`
	Type        *model.MangaType   `json:"type"`
	Tags        []string           `json:"tags"`
	Author      *string            `json:"author"`
	Artist      *string            `json:"artist"`
	Category    *string            `json:"category"`
}

// --- Responses ---

type MangaResponse struct {
	ID          uuid.UUID         `json:"id"`
	OwnerID     uuid.UUID         `json:"owner_id"`
	Title       string            `json:"title"`
	Slug        string            `json:"slug"`
	Description string            `json:"description"`
	CoverURL    string            `json:"cover_url"`
	Status      model.MangaStatus `json:"status"`
	Type        model.MangaType   `json:"type"`
	Tags        []string          `json:"tags"`
	Author      string            `json:"author"`
	Artist      string            `json:"artist"`
	Category    string            `json:"category"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewMangaResponse builds a MangaResponse. coverURL is passed separately because
// it is generated on-demand from the stored CoverKey (presigned URL, local crypto op).
func NewMangaResponse(m *model.Manga, coverURL string) MangaResponse {
	tags := []string(m.Tags)
	if tags == nil {
		tags = []string{}
	}
	return MangaResponse{
		ID:          m.ID,
		OwnerID:     m.OwnerID,
		Title:       m.Title,
		Slug:        m.Slug,
		Description: m.Description,
		CoverURL:    coverURL,
		Status:      m.Status,
		Type:        m.Type,
		Tags:        tags,
		Author:      m.Author,
		Artist:      m.Artist,
		Category:    m.Category,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
