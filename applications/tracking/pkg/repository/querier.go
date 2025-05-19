package repository

import (
	"context"
	"gorm.io/gorm"
)

type Querier interface {
	GetTrackingId(ctx context.Context, trackingId string) (*TrackingId, error)
	GetTrackingIds(ctx context.Context, filter *GetTrackingIdsFilter) ([]TrackingId, error)
}

type querierImpl struct {
	db *gorm.DB
}

func (q *querierImpl) GetTrackingId(ctx context.Context, trackingId string) (*TrackingId, error) {
	record := &TrackingId{}
	return record, q.db.Model(&TrackingId{}).Where("tracking_id = ?", trackingId).WithContext(ctx).First(&record).Error
}

func (q *querierImpl) GetTrackingIds(ctx context.Context, filter *GetTrackingIdsFilter) ([]TrackingId, error) {
	trackingIds := make([]TrackingId, 0)
	query := q.db.Model(&TrackingId{})
	if filter != nil {
		if len(filter.Ids) != 0 {
			query = query.Where("id in (?)", filter.Ids)
		}
	}
	return trackingIds, query.WithContext(ctx).Find(trackingIds).Error
}
