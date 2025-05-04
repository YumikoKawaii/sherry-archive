package repository

import "context"

type Querier interface {
	GetBooks(ctx context.Context, filter *GetBooksFilter) ([]Book, error)
	GetPages(ctx context.Context, filter *GetPagesFilter) ([]Page, error)
	GetAuthors(ctx context.Context, filter *GetAuthorsFilter) ([]Author, error)
	GetPublishers(ctx context.Context, filter *GetPublishersFilter) ([]Publisher, error)
}
