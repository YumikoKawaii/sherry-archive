package dto

import "github.com/google/uuid"

// --- Requests ---

type ReorderPagesRequest struct {
	PageIDs []uuid.UUID `json:"page_ids" binding:"required,min=1"`
}

// --- Responses ---

type PageUploadResponse struct {
	ID        uuid.UUID `json:"id"`
	Number    int       `json:"number"`
	ObjectKey string    `json:"object_key"`
}
