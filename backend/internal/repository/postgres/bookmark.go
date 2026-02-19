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

type BookmarkRepo struct{ db *sqlx.DB }

func NewBookmarkRepo(db *sqlx.DB) *BookmarkRepo { return &BookmarkRepo{db: db} }

func (r *BookmarkRepo) Upsert(ctx context.Context, b *model.Bookmark) error {
	const q = `
		INSERT INTO bookmarks (id, user_id, manga_id, chapter_id, last_page_number, created_at, updated_at)
		VALUES (:id, :user_id, :manga_id, :chapter_id, :last_page_number, :created_at, :updated_at)
		ON CONFLICT (user_id, manga_id) DO UPDATE SET
			chapter_id = EXCLUDED.chapter_id,
			last_page_number = EXCLUDED.last_page_number,
			updated_at = EXCLUDED.updated_at`
	_, err := r.db.NamedExecContext(ctx, q, b)
	return err
}

func (r *BookmarkRepo) GetByUserAndManga(ctx context.Context, userID, mangaID uuid.UUID) (*model.Bookmark, error) {
	var b model.Bookmark
	err := r.db.GetContext(ctx, &b,
		`SELECT * FROM bookmarks WHERE user_id = $1 AND manga_id = $2`, userID, mangaID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &b, err
}

func (r *BookmarkRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Bookmark, error) {
	var rows []*model.Bookmark
	err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM bookmarks WHERE user_id = $1 ORDER BY updated_at DESC`, userID)
	return rows, err
}

func (r *BookmarkRepo) Delete(ctx context.Context, userID, mangaID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM bookmarks WHERE user_id = $1 AND manga_id = $2`, userID, mangaID)
	return err
}
