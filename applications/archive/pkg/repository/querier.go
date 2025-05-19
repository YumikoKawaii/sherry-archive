package repository

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Querier interface {
	GetDocuments(ctx context.Context, filter *GetDocumentsFilter) ([]Document, error)
	UpsertDocument(ctx context.Context, book *Document) error
	GetChapters(ctx context.Context, filter *GetChaptersFilter) ([]Chapter, error)
	UpsertChapter(ctx context.Context, chapter *Chapter) error
	GetPages(ctx context.Context, filter *GetPagesFilter) ([]Page, error)
	UpsertPage(ctx context.Context, page *Page) error
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

func (q *querierImpl) GetDocuments(ctx context.Context, filter *GetDocumentsFilter) ([]Document, error) {
	queryBuilder := q.db.Model(&Document{})
	if filter != nil {
		if len(filter.Ids) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.Ids)
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
	documents := make([]Document, 0)
	return documents, queryBuilder.WithContext(ctx).Find(documents).Error
}

func (q *querierImpl) UpsertDocument(ctx context.Context, document *Document) error {
	return q.db.Model(&Document{}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).WithContext(ctx).Create(document).Error
}

func (q *querierImpl) GetChapters(ctx context.Context, filter *GetChaptersFilter) ([]Chapter, error) {
	query := q.db.Model(&Chapter{})
	if filter != nil {
		if len(filter.Ids) != 0 {
			query = query.Where("id in (?)", filter.Ids)
		}

		if filter.DocumentId != nil {
			query = query.Where("document_id = ?", filter.DocumentId)
		}

		if filter.Pagination != nil {
			offset := int((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
			query = query.Limit(int(filter.Pagination.PageSize)).Offset(offset)
		}
	}
	chapters := make([]Chapter, 0)
	return chapters, query.WithContext(ctx).Find(chapters).Error
}

func (q *querierImpl) UpsertChapter(ctx context.Context, chapter *Chapter) error {
	return q.db.Model(&Chapter{}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).WithContext(ctx).Create(chapter).Error
}

func (q *querierImpl) GetPages(ctx context.Context, filter *GetPagesFilter) ([]Page, error) {
	queryBuilder := q.db.Model(&Page{})
	if filter != nil {
		if len(filter.Ids) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.Ids)
		}

		if filter.ChapterId != nil {
			queryBuilder = queryBuilder.Where("chapter_id = ?", filter.ChapterId)
		}

		if filter.Pagination != nil {
			offset := int((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
			queryBuilder = queryBuilder.Limit(int(filter.Pagination.PageSize)).Offset(offset)
		}
	}
	pages := make([]Page, 0)
	return pages, queryBuilder.WithContext(ctx).Find(pages).Error
}

func (q *querierImpl) UpsertPage(ctx context.Context, page *Page) error {
	return q.db.Model(&Page{}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).WithContext(ctx).Create(page).Error
}

func (q *querierImpl) GetAuthors(ctx context.Context, filter *GetAuthorsFilter) ([]Author, error) {
	queryBuilder := q.db.Model(&Author{})
	if filter != nil {
		if len(filter.Ids) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.Ids)
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
		if len(filter.Ids) != 0 {
			queryBuilder = queryBuilder.Where("id in (?)", filter.Ids)
		}

		if filter.Pagination != nil {
			offset := int((filter.Pagination.Page - 1) * filter.Pagination.PageSize)
			queryBuilder = queryBuilder.Limit(int(filter.Pagination.PageSize)).Offset(offset)
		}
	}
	publishers := make([]Publisher, 0)
	return publishers, queryBuilder.Scan(publishers).Error
}
