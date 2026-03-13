package model

import (
	"time"

	"github.com/google/uuid"
)

type UploadTaskType string
type UploadTaskStatus string

const (
	UploadTaskTypeZip        UploadTaskType = "zip"
	UploadTaskTypeOneshotZip UploadTaskType = "oneshot_zip"

	UploadTaskStatusPending    UploadTaskStatus = "pending"
	UploadTaskStatusProcessing UploadTaskStatus = "processing"
	UploadTaskStatusDone       UploadTaskStatus = "done"
	UploadTaskStatusFailed     UploadTaskStatus = "failed"
)

type UploadTask struct {
	ID        uuid.UUID        `db:"id"`
	Type      UploadTaskType   `db:"type"`
	Status    UploadTaskStatus `db:"status"`
	OwnerID   uuid.UUID        `db:"owner_id"`
	MangaID   uuid.UUID        `db:"manga_id"`
	ChapterID uuid.NullUUID    `db:"chapter_id"` // NULL for oneshot_zip until Lambda creates the chapter
	S3Key     string           `db:"s3_key"`
	Error     string           `db:"error"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}
