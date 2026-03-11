package dto

// PagedResponse is a generic paginated list envelope.
type PagedResponse[T any] struct {
	Items []T `json:"items"`
	Total int  `json:"total"`
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
}

// ErrorResponse is returned for all error responses.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Concrete paged response types for Swagger (swag does not support generics).

type PagedMangaResponse struct {
	Items []MangaResponse `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

type PagedCommentResponse struct {
	Items []CommentResponse `json:"items"`
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}
