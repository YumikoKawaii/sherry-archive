package tracking

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
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

	// Build a single multi-row INSERT
	const cols = 8
	placeholders := make([]string, len(rows))
	args := make([]any, 0, len(rows)*cols)

	for i, r := range rows {
		base := i * cols
		placeholders[i] = fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8)
		args = append(args,
			r.DeviceID, r.UserID, r.Event, r.Properties,
			r.Referrer, r.IPHash, r.UserAgent, r.CreatedAt,
		)
	}

	q := `INSERT INTO events (device_id, user_id, event, properties, referrer, ip_hash, user_agent, created_at) VALUES ` +
		strings.Join(placeholders, ",")

	_, err := s.db.ExecContext(ctx, q, args...)
	return err
}
