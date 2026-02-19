package model

import (
	"time"

	"github.com/google/uuid"
)

type Chapter struct {
	ID        uuid.UUID `db:"id"`
	MangaID   uuid.UUID `db:"manga_id"`
	Number    float64   `db:"number"`
	Title     string    `db:"title"`
	PageCount int       `db:"page_count"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
