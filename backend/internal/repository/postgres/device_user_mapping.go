package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
)

type DeviceUserMappingRepo struct{ db *sqlx.DB }

func NewDeviceUserMappingRepo(db *sqlx.DB) *DeviceUserMappingRepo {
	return &DeviceUserMappingRepo{db: db}
}

func (r *DeviceUserMappingRepo) Upsert(ctx context.Context, deviceID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO device_user_mappings (device_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (device_id, user_id) DO NOTHING`,
		deviceID, userID,
	)
	return err
}

func (r *DeviceUserMappingRepo) GetUserByDevice(ctx context.Context, deviceID uuid.UUID) (uuid.UUID, error) {
	var userID uuid.UUID
	err := r.db.GetContext(ctx, &userID,
		`SELECT user_id FROM device_user_mappings WHERE device_id = $1 LIMIT 1`, deviceID)
	if err != nil {
		return uuid.Nil, apperror.ErrNotFound
	}
	return userID, nil
}
