package model

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	MangaID   uuid.UUID  `db:"manga_id"`
	ChapterID *uuid.UUID `db:"chapter_id"`
	Content   string     `db:"content"`
	Edited    bool       `db:"edited"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

// CommentWithAuthor is the flat struct returned by JOIN queries.
type CommentWithAuthor struct {
	Comment
	AuthorUsername  string `db:"author_username"`
	AuthorAvatarURL string `db:"author_avatar_url"`
}
