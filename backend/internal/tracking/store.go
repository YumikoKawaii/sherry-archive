package tracking

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Store is the interface for persisting events.
// Swap implementation to migrate to ClickHouse or a separate DB.
type Store interface {
	Insert(ctx context.Context, rows []EventRow) error
}

// EventRow is the fully enriched event ready to be persisted.
type EventRow struct {
	DeviceID   uuid.UUID
	UserID     *uuid.UUID
	Event      string
	Properties json.RawMessage
	Referrer   string
	IPHash     string
	UserAgent  string
	CreatedAt  time.Time
}

func HashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", h)
}

// PostgresStore persists events to the events table.
type PostgresStore struct{ db *sqlx.DB }

func NewPostgresStore(db *sqlx.DB) *PostgresStore { return &PostgresStore{db: db} }

func (s *PostgresStore) Insert(ctx context.Context, rows []EventRow) error {
	if len(rows) == 0 {
		return nil
	}
	const q = `
		INSERT INTO events (device_id, user_id, event, properties, referrer, ip_hash, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	for _, r := range rows {
		if _, err := s.db.ExecContext(ctx, q,
			r.DeviceID, r.UserID, r.Event, r.Properties,
			r.Referrer, r.IPHash, r.UserAgent, r.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}
