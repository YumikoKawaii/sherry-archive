package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SeenMangaRepo struct{ db *sqlx.DB }

func NewSeenMangaRepo(db *sqlx.DB) *SeenMangaRepo {
	return &SeenMangaRepo{db: db}
}

func (r *SeenMangaRepo) Add(ctx context.Context, identityID, mangaID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO seen_manga (identity_id, manga_id)
		VALUES ($1, $2)
		ON CONFLICT (identity_id, manga_id) DO NOTHING`,
		identityID, mangaID,
	)
	return err
}

func (r *SeenMangaRepo) ListIDsByIdentity(ctx context.Context, identityID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT manga_id FROM seen_manga WHERE identity_id = $1`, identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}

func (r *SeenMangaRepo) MergeInto(ctx context.Context, fromID, toID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO seen_manga (identity_id, manga_id, seen_at)
		SELECT $2, manga_id, seen_at FROM seen_manga WHERE identity_id = $1
		ON CONFLICT (identity_id, manga_id) DO NOTHING`,
		fromID, toID,
	)
	return err
}
