package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

type UserInterestRepo struct{ db *sqlx.DB }

func NewUserInterestRepo(db *sqlx.DB) *UserInterestRepo {
	return &UserInterestRepo{db: db}
}

func (r *UserInterestRepo) ListByIdentity(ctx context.Context, identityID uuid.UUID) ([]*model.UserInterest, error) {
	var rows []*model.UserInterest
	err := r.db.SelectContext(ctx, &rows,
		`SELECT identity_id, dimension, score, updated_at
		 FROM user_interests WHERE identity_id = $1`, identityID)
	return rows, err
}

func (r *UserInterestRepo) MergeInto(ctx context.Context, fromID, toID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_interests (identity_id, dimension, score, updated_at)
		SELECT $2, dimension, score, updated_at FROM user_interests WHERE identity_id = $1
		ON CONFLICT (identity_id, dimension) DO UPDATE
		  SET score      = GREATEST(EXCLUDED.score, user_interests.score),
		      updated_at = GREATEST(EXCLUDED.updated_at, user_interests.updated_at)`,
		fromID, toID,
	)
	return err
}

func (r *UserInterestRepo) UpsertBatch(ctx context.Context, interests []*model.UserInterest) error {
	if len(interests) == 0 {
		return nil
	}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO user_interests (identity_id, dimension, score, updated_at)
		VALUES (:identity_id, :dimension, :score, :updated_at)
		ON CONFLICT (identity_id, dimension) DO UPDATE
		  SET score = EXCLUDED.score, updated_at = EXCLUDED.updated_at`,
		interests,
	)
	return err
}

type InterestSyncWatermarkRepo struct{ db *sqlx.DB }

func NewInterestSyncWatermarkRepo(db *sqlx.DB) *InterestSyncWatermarkRepo {
	return &InterestSyncWatermarkRepo{db: db}
}

func (r *InterestSyncWatermarkRepo) Get(ctx context.Context, identityID uuid.UUID) (*model.InterestSyncWatermark, error) {
	var w model.InterestSyncWatermark
	err := r.db.GetContext(ctx, &w,
		`SELECT identity_id, last_synced_at FROM interest_sync_watermarks WHERE identity_id = $1`, identityID)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *InterestSyncWatermarkRepo) Upsert(ctx context.Context, identityID uuid.UUID, lastSyncedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO interest_sync_watermarks (identity_id, last_synced_at)
		VALUES ($1, $2)
		ON CONFLICT (identity_id) DO UPDATE SET last_synced_at = EXCLUDED.last_synced_at`,
		identityID, lastSyncedAt,
	)
	return err
}
