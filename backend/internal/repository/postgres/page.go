package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

type PageRepo struct{ db *sqlx.DB }

func NewPageRepo(db *sqlx.DB) *PageRepo { return &PageRepo{db: db} }

func (r *PageRepo) CreateBatch(ctx context.Context, pages []*model.Page) error {
	if len(pages) == 0 {
		return nil
	}
	const cols = 7
	placeholders := make([]string, len(pages))
	args := make([]any, 0, len(pages)*cols)
	for i, p := range pages {
		base := i * cols
		placeholders[i] = fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7)
		args = append(args, p.ID, p.ChapterID, p.Number, p.ObjectKey, p.Width, p.Height, p.CreatedAt)
	}
	q := fmt.Sprintf(`INSERT INTO pages (id, chapter_id, number, object_key, width, height, created_at) VALUES %s`,
		strings.Join(placeholders, ","))
	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}

func (r *PageRepo) GetByChapter(ctx context.Context, chapterID uuid.UUID) ([]*model.Page, error) {
	var rows []*model.Page
	err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM pages WHERE chapter_id = $1 ORDER BY number ASC`, chapterID)
	return rows, err
}

func (r *PageRepo) GetByChapterAndNumber(ctx context.Context, chapterID uuid.UUID, number int) (*model.Page, error) {
	var p model.Page
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM pages WHERE chapter_id = $1 AND number = $2`, chapterID, number)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &p, err
}

func (r *PageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM pages WHERE id = $1`, id)
	return err
}

func (r *PageRepo) UpdateNumbers(ctx context.Context, chapterID uuid.UUID, pageIDs []uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range pageIDs {
		if _, err := tx.ExecContext(ctx,
			`UPDATE pages SET number = $1 WHERE id = $2 AND chapter_id = $3`,
			i+1, id, chapterID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PageRepo) CountByChapter(ctx context.Context, chapterID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM pages WHERE chapter_id = $1`, chapterID)
	return count, err
}
