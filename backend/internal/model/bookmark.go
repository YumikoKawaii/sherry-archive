package model

import (
	"time"

	"github.com/google/uuid"
)

type Bookmark struct {
	ID             uuid.UUID `db:"id"`
	UserID         uuid.UUID `db:"user_id"`
	MangaID        uuid.UUID `db:"manga_id"`
	ChapterID      uuid.UUID `db:"chapter_id"`
	LastPageNumber int       `db:"last_page_number"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
