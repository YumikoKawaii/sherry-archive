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

type ZipMetadataSuggestions struct {
	ChapterNumber *float64 `json:"chapter_number,omitempty"`
	ChapterTitle  string   `json:"chapter_title,omitempty"`
	Author        string   `json:"author,omitempty"`
	Artist        string   `json:"artist,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Category      string   `json:"category,omitempty"`
	Language      string   `json:"language,omitempty"`
}

type ZipUploadResponse struct {
	Pages               []PageUploadResponse    `json:"pages"`
	MetadataSuggestions *ZipMetadataSuggestions `json:"metadata_suggestions,omitempty"`
}

type OneshotUploadResponse struct {
	Chapter             ChapterResponse         `json:"chapter"`
	Pages               []PageUploadResponse    `json:"pages"`
	MetadataSuggestions *ZipMetadataSuggestions `json:"metadata_suggestions,omitempty"`
}
