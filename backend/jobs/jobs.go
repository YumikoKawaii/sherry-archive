package jobs

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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

	// Build stop tags set
	stopTags := make(map[string]struct{})
	if cfg.Analytics.StopTags != "" {
		for _, t := range strings.Split(cfg.Analytics.StopTags, ",") {
			if t = strings.TrimSpace(t); t != "" {
				stopTags[t] = struct{}{}
			}
		}
	}

	ctx := context.Background()
	log.Println("==> Starting interest aggregation job")
	log.Printf("  stop tags: %v", cfg.Analytics.StopTags)

	if err := runAggregation(ctx, db, rdb, stopTags, cfg.Analytics.ContributionCap); err != nil {
		log.Fatalf("aggregation failed: %v", err)
	}

	log.Println("==> Done")
}

func runAggregation(ctx context.Context, db *sqlx.DB, rdb *redis.Client, stopTags map[string]struct{}, cap float64) error {
	mappingRepo := postgres.NewDeviceUserMappingRepo(db)
	interestRepo := postgres.NewUserInterestRepo(db)
	watermarkRepo := postgres.NewInterestSyncWatermarkRepo(db)

	// Popularity deltas accumulated across all devices: manga_id → delta
	popularityDeltas := make(map[string]float64)

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

		if err := processIdentity(ctx, db, rdb, interestRepo, watermarkRepo, deviceID, identityID, since, jobTime, stopTags, cap, popularityDeltas); err != nil {
			log.Printf("  [WARN] identity %s: %v", identityID, err)
			continue
		}
	}

	// Upsert popularity scores
	if len(popularityDeltas) > 0 {
		log.Printf("  upserting popularity for %d manga(s)", len(popularityDeltas))
		if err := upsertPopularity(ctx, db, popularityDeltas); err != nil {
			log.Printf("  [WARN] upsert popularity: %v", err)
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
	stopTags map[string]struct{},
	cap float64,
	popularityDeltas map[string]float64,
) error {
	// Fetch new events for this device since watermark
	events, err := fetchEventsSince(ctx, db, deviceID, since)
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}
	if len(events) == 0 {
		return nil
	}

	log.Printf("    device %s → identity %s: %d event(s)", deviceID, identityID, len(events))

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

	// Popularity: track capped contribution per manga per day for this device
	// key: manga_id+"_"+date, value: accumulated points in that day
	type dayKey struct {
		mangaID string
		day     string
	}
	dayContrib := make(map[dayKey]float64)

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

		// Popularity: apply cap per device per manga per day
		if trendPts, ok := trendingPointsMap[e.Event]; ok {
			dk := dayKey{mangaID, e.CreatedAt.Format("2006-01-02")}
			remaining := cap - dayContrib[dk]
			if remaining > 0 {
				allowed := trendPts
				if allowed > remaining {
					allowed = remaining
				}
				dayContrib[dk] += allowed
				popularityDeltas[mangaID] += allowed
			}
		}

		// Interest profile — skip stop tags
		activeTags := make([]string, 0, len(meta.Tags))
		for _, tag := range meta.Tags {
			if _, stopped := stopTags[tag]; !stopped {
				activeTags = append(activeTags, tag)
			}
		}
		if len(activeTags) > 0 {
			tagPts := pts / float64(len(activeTags))
			for _, tag := range activeTags {
				dim := "tag:" + tag
				scores[dim] = scores[dim]*interestDecay + tagPts
			}
		}
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
	CreatedAt  time.Time       `db:"created_at"`
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
			`SELECT event, properties, created_at FROM events WHERE device_id = $1 ORDER BY created_at ASC`, deviceID)
	} else {
		err = db.SelectContext(ctx, &rows,
			`SELECT event, properties, created_at FROM events WHERE device_id = $1 AND created_at > $2 ORDER BY created_at ASC`, deviceID, since)
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

var trendingPointsMap = map[string]float64{
	"manga_view":       1,
	"chapter_open":     3,
	"chapter_complete": 5,
}

func upsertPopularity(ctx context.Context, db *sqlx.DB, deltas map[string]float64) error {
	now := time.Now()
	for mangaIDStr, delta := range deltas {
		mangaID, err := uuid.Parse(mangaIDStr)
		if err != nil {
			continue
		}
		_, err = db.ExecContext(ctx, `
			INSERT INTO manga_popularity (manga_id, score, updated_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (manga_id) DO UPDATE
			  SET score = manga_popularity.score + EXCLUDED.score,
			      updated_at = EXCLUDED.updated_at`,
			mangaID, delta, now,
		)
		if err != nil {
			log.Printf("  [WARN] popularity upsert for %s: %v", mangaIDStr, err)
		}
	}
	return nil
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

