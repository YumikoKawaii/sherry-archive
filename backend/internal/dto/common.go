package dto

// PagedResponse is a generic paginated list envelope.
type PagedResponse[T any] struct {
	Items []T `json:"items"`
	Total int  `json:"total"`
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
}
