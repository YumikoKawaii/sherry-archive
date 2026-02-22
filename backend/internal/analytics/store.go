package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/tracking"
)

const (
	trendingKey        = "trending"
	seenPrefix         = "seen:"
	interestsPrefix    = "interests:"
	mangaMetaPrefix    = "manga:meta:"
	mangaMetaTTL       = time.Hour
	seenTTL            = 30 * 24 * time.Hour
)

// Store updates and queries the Redis-backed real-time analytics data.
type Store struct {
	rdb *redis.Client
	db  *sqlx.DB

	// Lua scripts — loaded once, referenced by SHA.
	updateInterest *redis.Script
	decayTrending  *redis.Script
}

func NewStore(rdb *redis.Client, db *sqlx.DB) *Store {
	s := &Store{rdb: rdb, db: db}
	// Atomic interest update with per-write decay.
	// KEYS[1] = interests:{device_id}, ARGV[1] = field, ARGV[2] = decay, ARGV[3] = points
	s.updateInterest = redis.NewScript(`
		local cur = redis.call('HGET', KEYS[1], ARGV[1])
		local score
		if cur then
			score = tonumber(cur) * tonumber(ARGV[2]) + tonumber(ARGV[3])
		else
			score = tonumber(ARGV[3])
		end
		if score <= 0 then
			redis.call('HDEL', KEYS[1], ARGV[1])
		else
			redis.call('HSET', KEYS[1], ARGV[1], tostring(score))
		end
		return tostring(score)
	`)
	// Decay all members of the trending sorted set.
	// KEYS[1] = trending, ARGV[1] = decay factor
	s.decayTrending = redis.NewScript(`
		local items = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', '+inf', 'WITHSCORES')
		for i = 1, #items, 2 do
			local newScore = tonumber(items[i+1]) * tonumber(ARGV[1])
			if newScore < 0.01 then
				redis.call('ZREM', KEYS[1], items[i])
			else
				redis.call('ZADD', KEYS[1], newScore, items[i])
			end
		end
		return 1
	`)
	return s
}

// --- Enricher (tracking.Enricher implementation) ---

// ProcessEvents is called by the tracking handler after events are stored.
// It updates trending scores and user interest profiles in Redis.
func (s *Store) ProcessEvents(ctx context.Context, events []tracking.EventRow) {
	for _, e := range events {
		mangaID := extractMangaID(e)
		if mangaID == "" {
			continue
		}

		// Update trending sorted set
		if pts, ok := trendingPoints[e.Event]; ok {
			s.rdb.ZIncrBy(ctx, trendingKey, pts, mangaID)
		}

		// Update interest profile and seen set
		if pts, ok := interestPoints[e.Event]; ok {
			deviceKey := seenPrefix + e.DeviceID.String()
			s.rdb.SAdd(ctx, deviceKey, mangaID)
			s.rdb.Expire(ctx, deviceKey, seenTTL)
			s.updateInterestProfile(ctx, e.DeviceID.String(), mangaID, pts)
		}
	}
}

func (s *Store) updateInterestProfile(ctx context.Context, deviceID, mangaID string, pts float64) {
	meta, err := s.getMangaMeta(ctx, mangaID)
	if err != nil || meta == nil {
		return
	}

	interestKey := interestsPrefix + deviceID

	// Tags — proportional split
	if len(meta.Tags) > 0 {
		tagPts := pts / float64(len(meta.Tags))
		for _, tag := range meta.Tags {
			s.updateInterest.Run(ctx, s.rdb, []string{interestKey},
				"tag:"+tag, interestDecay, tagPts)
		}
	}

	// Author and Category — full points (single-valued, no split)
	if meta.Author != "" {
		s.updateInterest.Run(ctx, s.rdb, []string{interestKey},
			"author:"+meta.Author, interestDecay, pts)
	}
	if meta.Category != "" {
		s.updateInterest.Run(ctx, s.rdb, []string{interestKey},
			"category:"+meta.Category, interestDecay, pts)
	}
}

// --- Manga metadata cache ---

type mangaMeta struct {
	Tags     []string
	Author   string
	Category string
}

func (s *Store) getMangaMeta(ctx context.Context, mangaID string) (*mangaMeta, error) {
	cacheKey := mangaMetaPrefix + mangaID

	// Try Redis cache first
	vals, err := s.rdb.HGetAll(ctx, cacheKey).Result()
	if err == nil && len(vals) > 0 {
		var tags []string
		_ = json.Unmarshal([]byte(vals["tags"]), &tags)
		return &mangaMeta{
			Tags:     tags,
			Author:   vals["author"],
			Category: vals["category"],
		}, nil
	}

	// Cache miss — query Postgres
	var row struct {
		Tags     pq.StringArray `db:"tags"`
		Author   string         `db:"author"`
		Category string         `db:"category"`
	}
	err = s.db.GetContext(ctx, &row,
		`SELECT tags, author, category FROM mangas WHERE id = $1`, mangaID)
	if err != nil {
		return nil, err
	}

	tagsJSON, _ := json.Marshal([]string(row.Tags))
	s.rdb.HSet(ctx, cacheKey,
		"tags", string(tagsJSON),
		"author", row.Author,
		"category", row.Category,
	)
	s.rdb.Expire(ctx, cacheKey, mangaMetaTTL)

	return &mangaMeta{
		Tags:     row.Tags,
		Author:   row.Author,
		Category: row.Category,
	}, nil
}

// --- Trending ---

type TrendingResult struct {
	Manga *model.Manga
	Score float64
}

func (s *Store) GetTrending(ctx context.Context, limit int) ([]*TrendingResult, error) {
	items, err := s.rdb.ZRevRangeWithScores(ctx, trendingKey, 0, int64(limit-1)).Result()
	if err != nil || len(items) == 0 {
		return nil, err
	}

	ids := make([]string, len(items))
	scoreMap := make(map[string]float64, len(items))
	for i, z := range items {
		ids[i] = z.Member.(string)
		scoreMap[ids[i]] = z.Score
	}

	mangas, err := s.fetchMangasByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Preserve trending order
	mangaMap := make(map[string]*model.Manga, len(mangas))
	for _, m := range mangas {
		mangaMap[m.ID.String()] = m
	}

	results := make([]*TrendingResult, 0, len(ids))
	for _, id := range ids {
		if m, ok := mangaMap[id]; ok {
			results = append(results, &TrendingResult{Manga: m, Score: scoreMap[id]})
		}
	}
	return results, nil
}

// --- Suggestions ---

func (s *Store) GetSuggestions(ctx context.Context, deviceID string, limit int) ([]*model.Manga, error) {
	interestKey := interestsPrefix + deviceID
	raw, err := s.rdb.HGetAll(ctx, interestKey).Result()
	if err != nil || len(raw) == 0 {
		return nil, nil // no profile yet
	}

	// Parse and sort interests by score descending
	type dim struct {
		key   string
		score float64
	}
	dims := make([]dim, 0, len(raw))
	for k, v := range raw {
		var score float64
		fmt.Sscanf(v, "%f", &score)
		dims = append(dims, dim{k, score})
	}
	sort.Slice(dims, func(i, j int) bool { return dims[i].score > dims[j].score })

	// Extract top interests per dimension
	var topTags, topAuthors, topCategories []string
	for _, d := range dims {
		switch {
		case strings.HasPrefix(d.key, "tag:") && len(topTags) < 5:
			topTags = append(topTags, strings.TrimPrefix(d.key, "tag:"))
		case strings.HasPrefix(d.key, "author:") && len(topAuthors) < 3:
			topAuthors = append(topAuthors, strings.TrimPrefix(d.key, "author:"))
		case strings.HasPrefix(d.key, "category:") && len(topCategories) < 3:
			topCategories = append(topCategories, strings.TrimPrefix(d.key, "category:"))
		}
	}

	// Get seen manga IDs to exclude
	seenStrs, _ := s.rdb.SMembers(ctx, seenPrefix+deviceID).Result()
	seenIDs := make([]uuid.UUID, 0, len(seenStrs))
	for _, s := range seenStrs {
		if id, err := uuid.Parse(s); err == nil {
			seenIDs = append(seenIDs, id)
		}
	}

	return s.querySuggestions(ctx, topTags, topAuthors, topCategories, seenIDs, limit)
}

func (s *Store) querySuggestions(
	ctx context.Context,
	tags, authors, categories []string,
	seenIDs []uuid.UUID,
	limit int,
) ([]*model.Manga, error) {
	if len(tags) == 0 && len(authors) == 0 && len(categories) == 0 {
		return nil, nil
	}

	var mangas []*model.Manga
	err := s.db.SelectContext(ctx, &mangas, `
		SELECT * FROM mangas
		WHERE ($1::uuid[] IS NULL OR id != ALL($1::uuid[]))
		  AND (
		        tags && $2::text[]
		     OR (author != '' AND author = ANY($3::text[]))
		     OR (category != '' AND category = ANY($4::text[]))
		  )
		ORDER BY created_at DESC
		LIMIT $5`,
		pq.Array(seenIDs),
		pq.Array(tags),
		pq.Array(authors),
		pq.Array(categories),
		limit,
	)
	return mangas, err
}

// --- Decay loop ---

// StartDecay runs the hourly trending decay in the background.
// Call it as a goroutine from serve/server.go.
func (s *Store) StartDecay(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.decayTrending.Run(ctx, s.rdb, []string{trendingKey}, trendingDecay)
		case <-ctx.Done():
			return
		}
	}
}

// --- Helpers ---

func extractMangaID(e tracking.EventRow) string {
	var props map[string]any
	if err := json.Unmarshal(e.Properties, &props); err != nil {
		return ""
	}
	id, _ := props["manga_id"].(string)
	return id
}

func (s *Store) fetchMangasByIDs(ctx context.Context, ids []string) ([]*model.Manga, error) {
	uids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if u, err := uuid.Parse(id); err == nil {
			uids = append(uids, u)
		}
	}
	if len(uids) == 0 {
		return nil, nil
	}
	var mangas []*model.Manga
	err := s.db.SelectContext(ctx, &mangas,
		`SELECT * FROM mangas WHERE id = ANY($1)`, pq.Array(uids))
	return mangas, err
}
