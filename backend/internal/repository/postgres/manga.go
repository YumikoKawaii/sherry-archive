package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/pkg/pagination"
)

type MangaRepo struct{ db *sqlx.DB }

func NewMangaRepo(db *sqlx.DB) *MangaRepo { return &MangaRepo{db: db} }

func (r *MangaRepo) Create(ctx context.Context, m *model.Manga) error {
	const q = `
		INSERT INTO mangas (id, owner_id, title, slug, description, cover_key, status, tags, created_at, updated_at)
		VALUES (:id, :owner_id, :title, :slug, :description, :cover_key, :status, :tags, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, m)
	return err
}

func (r *MangaRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Manga, error) {
	var m model.Manga
	err := r.db.GetContext(ctx, &m, `SELECT * FROM mangas WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &m, err
}

func (r *MangaRepo) GetBySlug(ctx context.Context, slug string) (*model.Manga, error) {
	var m model.Manga
	err := r.db.GetContext(ctx, &m, `SELECT * FROM mangas WHERE slug = $1`, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &m, err
}

func (r *MangaRepo) List(ctx context.Context, filter repository.MangaFilter, p pagination.Params) ([]*model.Manga, int, error) {
	where, args := buildMangaWhere(filter)
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM mangas %s`, where)
	var total int
	if err := r.db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, err
	}

	order := mangaOrderBy(filter.Sort)
	dataQ := fmt.Sprintf(`SELECT * FROM mangas %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, order, len(args)+1, len(args)+2)
	args = append(args, p.Limit, p.Offset)

	var rows []*model.Manga
	if err := r.db.SelectContext(ctx, &rows, dataQ, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *MangaRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID, p pagination.Params) ([]*model.Manga, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM mangas WHERE owner_id = $1`, ownerID); err != nil {
		return nil, 0, err
	}
	var rows []*model.Manga
	err := r.db.SelectContext(ctx, &rows,
		`SELECT * FROM mangas WHERE owner_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		ownerID, p.Limit, p.Offset)
	return rows, total, err
}

func (r *MangaRepo) Update(ctx context.Context, m *model.Manga) error {
	const q = `
		UPDATE mangas SET title=:title, slug=:slug, description=:description, cover_key=:cover_key,
		status=:status, tags=:tags, updated_at=:updated_at WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, m)
	return err
}

func (r *MangaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM mangas WHERE id = $1`, id)
	return err
}

func (r *MangaRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM mangas WHERE slug = $1)`, slug)
	return exists, err
}

func buildMangaWhere(f repository.MangaFilter) (string, []any) {
	var clauses []string
	var args []any
	idx := 1

	if f.Query != "" {
		clauses = append(clauses, fmt.Sprintf(`(title ILIKE $%d OR description ILIKE $%d)`, idx, idx+1))
		like := "%" + f.Query + "%"
		args = append(args, like, like)
		idx += 2
	}
	if f.Status != "" {
		clauses = append(clauses, fmt.Sprintf(`status = $%d`, idx))
		args = append(args, f.Status)
		idx++
	}
	if len(f.Tags) > 0 {
		clauses = append(clauses, fmt.Sprintf(`tags @> $%d`, idx))
		args = append(args, pq.Array(f.Tags))
		idx++
	}

	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func mangaOrderBy(sort string) string {
	switch sort {
	case "oldest":
		return "created_at ASC"
	case "title":
		return "title ASC"
	default:
		return "created_at DESC"
	}
}
