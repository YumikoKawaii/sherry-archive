package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

// EnqueueResponse is returned immediately on a zip upload — the client polls TaskID for status.
type EnqueueResponse struct {
	TaskID uuid.UUID `json:"task_id"`
}

type UploadTaskResponse struct {
	ID        uuid.UUID              `json:"id"`
	Type      model.UploadTaskType   `json:"type"`
	Status    model.UploadTaskStatus `json:"status"`
	MangaID   uuid.UUID              `json:"manga_id"`
	ChapterID *uuid.UUID             `json:"chapter_id,omitempty"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func NewUploadTaskResponse(t *model.UploadTask) UploadTaskResponse {
	resp := UploadTaskResponse{
		ID:        t.ID,
		Type:      t.Type,
		Status:    t.Status,
		MangaID:   t.MangaID,
		Error:     t.Error,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if t.ChapterID.Valid {
		resp.ChapterID = &t.ChapterID.UUID
	}
	return resp
}
