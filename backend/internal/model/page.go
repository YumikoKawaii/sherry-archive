package model

import (
	"time"

	"github.com/google/uuid"
)

type Page struct {
	ID        uuid.UUID `db:"id"`
	ChapterID uuid.UUID `db:"chapter_id"`
	Number    int       `db:"number"`
	ObjectKey string    `db:"object_key"`
	Width     int       `db:"width"`
	Height    int       `db:"height"`
	CreatedAt time.Time `db:"created_at"`
}
