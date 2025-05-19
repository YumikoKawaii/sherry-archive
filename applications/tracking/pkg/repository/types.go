package repository

import "time"

type TrackingId struct {
	Id        string
	CreatedAt time.Time  `gorm:"created_at"`
	UpdatedAt time.Time  `gorm:"updated_at"`
	DeletedAt *time.Time `gorm:"deleted_at"`
}

type GetTrackingIdsFilter struct {
	Ids []string
}
