package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/pkg/pagination"
)

type CommentRepo struct{ db *sqlx.DB }

func NewCommentRepo(db *sqlx.DB) *CommentRepo { return &CommentRepo{db: db} }

const commentJoin = `
	SELECT c.*, u.username AS author_username, u.avatar_url AS author_avatar_url
	FROM comments c
	JOIN users u ON u.id = c.user_id`

func (r *CommentRepo) Create(ctx context.Context, c *model.Comment) error {
	const q = `
		INSERT INTO comments (id, user_id, manga_id, chapter_id, content, edited, created_at, updated_at)
		VALUES (:id, :user_id, :manga_id, :chapter_id, :content, :edited, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, c)
	return err
}

func (r *CommentRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.CommentWithAuthor, error) {
	var c model.CommentWithAuthor
	err := r.db.GetContext(ctx, &c, commentJoin+` WHERE c.id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &c, err
}

func (r *CommentRepo) ListByManga(ctx context.Context, mangaID uuid.UUID, p pagination.Params) ([]*model.CommentWithAuthor, int, error) {
	var total int
	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM comments WHERE manga_id = $1 AND chapter_id IS NULL`, mangaID)
	if err != nil {
		return nil, 0, err
	}

	var rows []*model.CommentWithAuthor
	err = r.db.SelectContext(ctx, &rows,
		commentJoin+` WHERE c.manga_id = $1 AND c.chapter_id IS NULL
		ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`,
		mangaID, p.Limit, p.Offset)
	return rows, total, err
}

func (r *CommentRepo) ListByChapter(ctx context.Context, chapterID uuid.UUID, p pagination.Params) ([]*model.CommentWithAuthor, int, error) {
	var total int
	err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM comments WHERE chapter_id = $1`, chapterID)
	if err != nil {
		return nil, 0, err
	}

	var rows []*model.CommentWithAuthor
	err = r.db.SelectContext(ctx, &rows,
		commentJoin+` WHERE c.chapter_id = $1
		ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`,
		chapterID, p.Limit, p.Offset)
	return rows, total, err
}

func (r *CommentRepo) Update(ctx context.Context, c *model.Comment) error {
	const q = `UPDATE comments SET content = :content, edited = :edited, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, q, c)
	return err
}

func (r *CommentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM comments WHERE id = $1`, id)
	return err
}
