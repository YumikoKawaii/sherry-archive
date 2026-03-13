package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

type UploadTaskRepo struct{ db *sqlx.DB }

func NewUploadTaskRepo(db *sqlx.DB) *UploadTaskRepo { return &UploadTaskRepo{db: db} }

func (r *UploadTaskRepo) Create(ctx context.Context, t *model.UploadTask) error {
	const q = `
		INSERT INTO upload_tasks (id, type, status, owner_id, manga_id, chapter_id, s3_key, error, created_at, updated_at)
		VALUES (:id, :type, :status, :owner_id, :manga_id, :chapter_id, :s3_key, :error, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, t)
	return err
}

func (r *UploadTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.UploadTask, error) {
	var t model.UploadTask
	err := r.db.GetContext(ctx, &t, `SELECT * FROM upload_tasks WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &t, err
}

func (r *UploadTaskRepo) ClaimProcessing(ctx context.Context, id uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE upload_tasks SET status = $1, updated_at = $2 WHERE id = $3 AND status = $4`,
		model.UploadTaskStatusProcessing, time.Now(), id, model.UploadTaskStatusPending,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *UploadTaskRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status model.UploadTaskStatus, errMsg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE upload_tasks SET status = $1, error = $2, updated_at = $3 WHERE id = $4`,
		status, errMsg, time.Now(), id,
	)
	return err
}

func (r *UploadTaskRepo) SetChapterAndDone(ctx context.Context, id uuid.UUID, chapterID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE upload_tasks SET status = $1, chapter_id = $2, updated_at = $3 WHERE id = $4`,
		model.UploadTaskStatusDone, chapterID, time.Now(), id,
	)
	return err
}
