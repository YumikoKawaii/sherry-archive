package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

type ChapterRepo struct{ db *sqlx.DB }

func NewChapterRepo(db *sqlx.DB) *ChapterRepo { return &ChapterRepo{db: db} }

func (r *ChapterRepo) Create(ctx context.Context, ch *model.Chapter) error {
	const q = `
		INSERT INTO chapters (id, manga_id, number, title, page_count, created_at, updated_at)
		VALUES (:id, :manga_id, :number, :title, :page_count, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, ch)
	return err
}

func (r *ChapterRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Chapter, error) {
	var ch model.Chapter
	err := r.db.GetContext(ctx, &ch, `SELECT * FROM chapters WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &ch, err
}

func (r *ChapterRepo) GetByMangaAndNumber(ctx context.Context, mangaID uuid.UUID, number float64) (*model.Chapter, error) {
	var ch model.Chapter
	err := r.db.GetContext(ctx, &ch, `SELECT * FROM chapters WHERE manga_id = $1 AND number = $2`, mangaID, number)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &ch, err
}

func (r *ChapterRepo) ListByManga(ctx context.Context, mangaID uuid.UUID) ([]*model.Chapter, error) {
	var rows []*model.Chapter
	err := r.db.SelectContext(ctx, &rows, `SELECT * FROM chapters WHERE manga_id = $1 ORDER BY number ASC`, mangaID)
	return rows, err
}

func (r *ChapterRepo) Update(ctx context.Context, ch *model.Chapter) error {
	const q = `UPDATE chapters SET number=:number, title=:title, updated_at=:updated_at WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, ch)
	return err
}

func (r *ChapterRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chapters WHERE id = $1`, id)
	return err
}

func (r *ChapterRepo) UpdatePageCount(ctx context.Context, id uuid.UUID, count int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE chapters SET page_count = $1, updated_at = NOW() WHERE id = $2`, count, id)
	return err
}
