package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type MangaStatus string

const (
	StatusOngoing   MangaStatus = "ongoing"
	StatusCompleted MangaStatus = "completed"
	StatusHiatus    MangaStatus = "hiatus"
)

type MangaType string

const (
	TypeSeries  MangaType = "series"
	TypeOneshot MangaType = "oneshot"
)

type Manga struct {
	ID          uuid.UUID      `db:"id"`
	OwnerID     uuid.UUID      `db:"owner_id"`
	Title       string         `db:"title"`
	Slug        string         `db:"slug"`
	Description string         `db:"description"`
	CoverKey    string         `db:"cover_key"`
	Status      MangaStatus    `db:"status"`
	Type        MangaType      `db:"type"`
	Tags        pq.StringArray `db:"tags"`
	Author      string         `db:"author"`
	Artist      string         `db:"artist"`
	Category    string         `db:"category"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}
