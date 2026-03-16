package model

import (
	"time"

	"github.com/google/uuid"
)

type DeviceUserMapping struct {
	DeviceID  uuid.UUID `db:"device_id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type UserInterest struct {
	IdentityID uuid.UUID `db:"identity_id"`
	Dimension  string    `db:"dimension"`
	Score      float64   `db:"score"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type InterestSyncWatermark struct {
	IdentityID   uuid.UUID `db:"identity_id"`
	LastSyncedAt time.Time `db:"last_synced_at"`
}

type SeenManga struct {
	IdentityID uuid.UUID `db:"identity_id"`
	MangaID    uuid.UUID `db:"manga_id"`
	SeenAt     time.Time `db:"seen_at"`
}
