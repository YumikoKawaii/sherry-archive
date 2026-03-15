package jobs

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
)

const (
	interestDecay     = 0.98
	interestCacheTTL  = 24 * time.Hour
	interestsPrefix   = "interests:"
)

func Run(_ *cobra.Command, _ []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := postgres.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	rdb, err := connectRedis(cfg)
	if err != nil {
		log.Fatalf("redis connect: %v", err)
	}
	defer rdb.Close()

	ctx := context.Background()
	log.Println("==> Starting interest aggregation job")

	if err := runAggregation(ctx, db, rdb); err != nil {
		log.Fatalf("aggregation failed: %v", err)
	}

	log.Println("==> Done")
}

func runAggregation(ctx context.Context, db *sqlx.DB, rdb *redis.Client) error {
	mappingRepo := postgres.NewDeviceUserMappingRepo(db)
	interestRepo := postgres.NewUserInterestRepo(db)
	watermarkRepo := postgres.NewInterestSyncWatermarkRepo(db)

	// Find all distinct device_ids that have unprocessed events
	deviceIDs, err := fetchDevicesWithNewEvents(ctx, db, watermarkRepo)
	if err != nil {
		return fmt.Errorf("fetch devices: %w", err)
	}

	log.Printf("  processing %d device(s)", len(deviceIDs))
	jobTime := time.Now()

	for _, deviceID := range deviceIDs {
		// Resolve identity: use user_id if mapped, else device_id
		identityID := deviceID
		userID, err := mappingRepo.GetUserByDevice(ctx, deviceID)
		if err == nil {
			identityID = userID
		}

		watermark, err := watermarkRepo.Get(ctx, identityID)
		var since time.Time
		if err == nil {
			since = watermark.LastSyncedAt
		}

		if err := processIdentity(ctx, db, rdb, interestRepo, watermarkRepo, deviceID, identityID, since, jobTime); err != nil {
			log.Printf("  [WARN] identity %s: %v", identityID, err)
			continue
		}
	}
	return nil
}

func processIdentity(
	ctx context.Context,
	db *sqlx.DB,
	rdb *redis.Client,
	interestRepo *postgres.UserInterestRepo,
	watermarkRepo *postgres.InterestSyncWatermarkRepo,
	deviceID, identityID uuid.UUID,
	since, jobTime time.Time,
) error {
	// Fetch new events for this device since watermark
	events, err := fetchEventsSince(ctx, db, deviceID, since)
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}
	if len(events) == 0 {
		return nil
	}

	// Fetch manga metadata for all unique manga_ids in events
	mangaIDs := extractMangaIDs(events)
	metaMap, err := fetchMangaMeta(ctx, db, mangaIDs)
	if err != nil {
		return fmt.Errorf("fetch manga meta: %w", err)
	}

	// Load existing interests
	existing, err := interestRepo.ListByIdentity(ctx, identityID)
	if err != nil {
		return fmt.Errorf("load interests: %w", err)
	}
	scores := make(map[string]float64, len(existing))
	for _, row := range existing {
		scores[row.Dimension] = row.Score
	}

	// Calculate deltas from new events
	for _, e := range events {
		pts, ok := interestPointsMap[e.Event]
		if !ok {
			continue
		}
		mangaID := extractMangaIDFromProps(e.Properties)
		if mangaID == "" {
			continue
		}
		meta, ok := metaMap[mangaID]
		if !ok {
			continue
		}

		// Tags — proportional split
		if len(meta.Tags) > 0 {
			tagPts := pts / float64(len(meta.Tags))
			for _, tag := range meta.Tags {
				dim := "tag:" + tag
				scores[dim] = scores[dim]*interestDecay + tagPts
			}
		}
		// Author and category — full points
		if meta.Author != "" {
			dim := "author:" + meta.Author
			scores[dim] = scores[dim]*interestDecay + pts
		}
		if meta.Category != "" {
			dim := "category:" + meta.Category
			scores[dim] = scores[dim]*interestDecay + pts
		}
	}

	// Remove zero/negative scores
	interests := make([]*model.UserInterest, 0, len(scores))
	now := time.Now()
	for dim, score := range scores {
		if score > 0 {
			interests = append(interests, &model.UserInterest{
				IdentityID: identityID,
				Dimension:  dim,
				Score:      score,
				UpdatedAt:  now,
			})
		}
	}

	// Upsert to DB
	if err := interestRepo.UpsertBatch(ctx, interests); err != nil {
		return fmt.Errorf("upsert interests: %w", err)
	}

	// Populate Redis cache
	if err := populateCache(ctx, rdb, identityID.String(), interests); err != nil {
		log.Printf("  [WARN] redis cache for %s: %v", identityID, err)
	}

	// Update watermark
	return watermarkRepo.Upsert(ctx, identityID, jobTime)
}

func populateCache(ctx context.Context, rdb *redis.Client, identityID string, interests []*model.UserInterest) error {
	if len(interests) == 0 {
		return nil
	}
	cacheKey := interestsPrefix + identityID
	args := make([]interface{}, 0, len(interests)*2)
	for _, i := range interests {
		args = append(args, i.Dimension, fmt.Sprintf("%f", i.Score))
	}
	if err := rdb.Del(ctx, cacheKey).Err(); err != nil {
		return err
	}
	if err := rdb.HSet(ctx, cacheKey, args...).Err(); err != nil {
		return err
	}
	return rdb.Expire(ctx, cacheKey, interestCacheTTL).Err()
}

// --- DB helpers ---

type eventRow struct {
	Event      string          `db:"event"`
	Properties json.RawMessage `db:"properties"`
}

func fetchDevicesWithNewEvents(ctx context.Context, db *sqlx.DB, watermarkRepo *postgres.InterestSyncWatermarkRepo) ([]uuid.UUID, error) {
	var rows []struct {
		DeviceID uuid.UUID `db:"device_id"`
	}
	err := db.SelectContext(ctx, &rows, `
		SELECT DISTINCT e.device_id
		FROM events e
		LEFT JOIN interest_sync_watermarks w ON w.identity_id = e.device_id
		WHERE w.last_synced_at IS NULL OR e.created_at > w.last_synced_at
	`)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.DeviceID
	}
	return ids, nil
}

func fetchEventsSince(ctx context.Context, db *sqlx.DB, deviceID uuid.UUID, since time.Time) ([]eventRow, error) {
	var rows []eventRow
	var err error
	if since.IsZero() {
		err = db.SelectContext(ctx, &rows,
			`SELECT event, properties FROM events WHERE device_id = $1 ORDER BY created_at ASC`, deviceID)
	} else {
		err = db.SelectContext(ctx, &rows,
			`SELECT event, properties FROM events WHERE device_id = $1 AND created_at > $2 ORDER BY created_at ASC`, deviceID, since)
	}
	return rows, err
}

type mangaMeta struct {
	Tags     []string
	Author   string
	Category string
}

func fetchMangaMeta(ctx context.Context, db *sqlx.DB, mangaIDs []string) (map[string]*mangaMeta, error) {
	if len(mangaIDs) == 0 {
		return map[string]*mangaMeta{}, nil
	}
	uids := make([]uuid.UUID, 0, len(mangaIDs))
	for _, id := range mangaIDs {
		if u, err := uuid.Parse(id); err == nil {
			uids = append(uids, u)
		}
	}
	var rows []struct {
		ID       uuid.UUID      `db:"id"`
		Tags     pq.StringArray `db:"tags"`
		Author   string         `db:"author"`
		Category string         `db:"category"`
	}
	err := db.SelectContext(ctx, &rows,
		`SELECT id, tags, author, category FROM mangas WHERE id = ANY($1)`, pq.Array(uids))
	if err != nil {
		return nil, err
	}
	result := make(map[string]*mangaMeta, len(rows))
	for _, r := range rows {
		result[r.ID.String()] = &mangaMeta{
			Tags:     []string(r.Tags),
			Author:   r.Author,
			Category: r.Category,
		}
	}
	return result, nil
}

func extractMangaIDs(events []eventRow) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, e := range events {
		if id := extractMangaIDFromProps(e.Properties); id != "" {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func extractMangaIDFromProps(props json.RawMessage) string {
	var m map[string]interface{}
	if err := json.Unmarshal(props, &m); err != nil {
		return ""
	}
	id, _ := m["manga_id"].(string)
	return id
}

var interestPointsMap = map[string]float64{
	"manga_view":       1,
	"chapter_open":     3,
	"chapter_complete": 5,
	"comment_post":     4,
	"bookmark_add":     8,
	"bookmark_remove":  -3,
}

// connectRedis reuses the same logic as serve/server.go
func connectRedis(cfg *config.Application) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if cfg.Redis.TLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}

