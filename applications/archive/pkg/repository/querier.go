package repository

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Querier interface {
	GetBooks(ctx context.Context, filter *GetBooksFilter) ([]Book, error)
	UpsertBook(ctx context.Context, book *Book) error
	GetPages(ctx context.Context, filter *GetPagesFilter) ([]Page, error)
	GetAuthors(ctx context.Context, filter *GetAuthorsFilter) ([]Author, error)
	GetPublishers(ctx context.Context, filter *GetPublishersFilter) ([]Publisher, error)
}

func NewQuerier(db *gorm.DB) Querier {
	return &querierImpl{
		db: db,
	}
}

type querierImpl struct {
	db *gorm.DB
}

func (q *querierImpl) GetBooks(ctx context.Context, filter *GetBooksFilter) ([]Book, error) {
	queryBuilder := q.db.Model(&Book{})
	if filter != nil {
		if len(filter.IDs) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.IDs)
		}

		if filter.AuthorId != nil {
			queryBuilder = queryBuilder.Where("author_id = ?", filter.AuthorId)
		}

		if filter.CategoryId != nil {
			queryBuilder = queryBuilder.Where("category_id = ?", filter.CategoryId)
		}

		if filter.PublisherId != nil {
			queryBuilder = queryBuilder.Where("publisher_id = ?", filter.PublisherId)
		}

		if filter.Pagination != nil {
			offset := int((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
			queryBuilder = queryBuilder.Limit(int(filter.Pagination.PageSize)).Offset(offset)
		}
	}
	books := make([]Book, 0)
	return books, queryBuilder.Scan(books).Error
}

func (q *querierImpl) UpsertBook(ctx context.Context, book *Book) error {
	return q.db.Model(&Book{}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(book).Error
}

func (q *querierImpl) GetPages(ctx context.Context, filter *GetPagesFilter) ([]Page, error) {
	queryBuilder := q.db.Model(&Page{})
	if filter != nil {
		if len(filter.IDs) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.IDs)
		}

		if filter.BookId != nil {
			queryBuilder = queryBuilder.Where("book_id = ?", filter.BookId)
		}

		if filter.Pagination != nil {
			offset := int((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
			queryBuilder = queryBuilder.Limit(int(filter.Pagination.PageSize)).Offset(offset)
		}
	}
	pages := make([]Page, 0)
	return pages, queryBuilder.Scan(pages).Error
}

func (q *querierImpl) GetAuthors(ctx context.Context, filter *GetAuthorsFilter) ([]Author, error) {
	queryBuilder := q.db.Model(&Author{})
	if filter != nil {
		if len(filter.IDs) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.IDs)
		}

		if filter.Pagination != nil {
			offset := int((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
			queryBuilder = queryBuilder.Limit(int(filter.Pagination.PageSize)).Offset(offset)
		}
	}
	authors := make([]Author, 0)
	return authors, queryBuilder.Scan(authors).Error
}

func (q *querierImpl) GetPublishers(ctx context.Context, filter *GetPublishersFilter) ([]Publisher, error) {
	queryBuilder := q.db.Model(&Publisher{})
	if filter != nil {
		if len(filter.IDs) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.IDs)
		}

		if filter.Pagination != nil {
			offset := int((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
			queryBuilder = queryBuilder.Limit(int(filter.Pagination.PageSize)).Offset(offset)
		}
	}
	publishers := make([]Publisher, 0)
	return publishers, queryBuilder.Scan(publishers).Error
}
