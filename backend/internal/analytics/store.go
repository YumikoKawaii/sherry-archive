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
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/internal/tracking"
)

const (
	trendingKey        = "trending"
	interestsPrefix    = "interests:"
	mangaMetaPrefix    = "manga:meta:"
	contributedPrefix  = "contributed:"
	contributionWindow = 24 * time.Hour
	mangaMetaTTL       = time.Hour
)

// Store updates and queries the Redis-backed real-time analytics data.
type Store struct {
	rdb             *redis.Client
	db              *sqlx.DB
	seenRepo        repository.SeenMangaRepository
	contributionCap float64
	decayInterval   time.Duration
	stopTags        map[string]struct{}

	// Lua scripts — loaded once, referenced by SHA.
	decayTrending    *redis.Script
	contributePoints *redis.Script
}

func NewStore(rdb *redis.Client, db *sqlx.DB, seenRepo repository.SeenMangaRepository, contributionCap float64, decayInterval time.Duration, stopTags map[string]struct{}) *Store {
	s := &Store{rdb: rdb, db: db, seenRepo: seenRepo, contributionCap: contributionCap, decayInterval: decayInterval, stopTags: stopTags}
	// Cap per-device contribution to trending within the 24h window.
	// KEYS[1] = contributed:{device_id}:{manga_id}
	// ARGV[1] = requested points, ARGV[2] = cap, ARGV[3] = window TTL in seconds
	// Returns the actual points to add (0 if already at cap).
	s.contributePoints = redis.NewScript(`
		local current = tonumber(redis.call('GET', KEYS[1]) or 0)
		local cap = tonumber(ARGV[2])
		if current >= cap then
			return '0'
		end
		local pts = tonumber(ARGV[1])
		local allowed = math.min(pts, cap - current)
		local newval = current + allowed
		redis.call('SET', KEYS[1], tostring(newval))
		if current == 0 then
			redis.call('EXPIRE', KEYS[1], tonumber(ARGV[3]))
		end
		return tostring(allowed)
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

		// Update trending sorted set, capped per device per manga per window
		if pts, ok := trendingPoints[e.Event]; ok {
			contribKey := contributedPrefix + e.DeviceID.String() + ":" + mangaID
			res, err := s.contributePoints.Run(ctx, s.rdb, []string{contribKey},
				pts, s.contributionCap, int(contributionWindow.Seconds())).Text()
			if err == nil {
				var allowed float64
				fmt.Sscanf(res, "%f", &allowed)
				if allowed > 0 {
					s.rdb.ZIncrBy(ctx, trendingKey, allowed, mangaID)
				}
			}
		}

		// Record seen manga in DB (permanent exclusion from suggestions).
		// Use user_id when available so logged-in reads are immediately excluded
		// from user-scoped suggestions without waiting for a merge.
		if _, ok := interestPoints[e.Event]; ok {
			if mID, err := uuid.Parse(mangaID); err == nil {
				identityID := e.DeviceID
				if e.UserID != nil {
					identityID = *e.UserID
				}
				_ = s.seenRepo.Add(ctx, identityID, mID)
			}
		}
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

const interestCacheTTL = 24 * time.Hour

// GetSuggestions returns personalised manga recommendations.
// If userID is non-nil, the user's interest profile is used with fallback to deviceID.
// If contextMangaID is non-nil, the currently viewing manga's metadata boosts matching.
func (s *Store) GetSuggestions(ctx context.Context, userID *uuid.UUID, deviceID string, contextMangaID *uuid.UUID, limit int) ([]*model.Manga, error) {
	var identityID string
	if userID != nil {
		identityID = userID.String()
	} else {
		identityID = deviceID
	}

	dims, err := s.loadInterests(ctx, identityID)
	if err != nil || len(dims) == 0 {
		// Fallback to device profile if user profile is empty
		if userID != nil && deviceID != "" {
			dims, err = s.loadInterests(ctx, deviceID)
		}
	}

	// Cold-start fallback: no interest profile built yet
	if len(dims) == 0 {
		var identityUUID uuid.UUID
		if id, err := uuid.Parse(identityID); err == nil {
			identityUUID = id
		}
		seenIDs, _ := s.seenRepo.ListIDsByIdentity(ctx, identityUUID)
		if contextMangaID != nil {
			seenIDs = append(seenIDs, *contextMangaID)
		}
		return s.coldStartSuggestions(ctx, contextMangaID, seenIDs, limit)
	}

	// Filter stop tags from dims
	filtered := dims[:0]
	for _, d := range dims {
		if strings.HasPrefix(d.key, "tag:") {
			tag := strings.TrimPrefix(d.key, "tag:")
			if _, stopped := s.stopTags[tag]; stopped {
				continue
			}
		}
		filtered = append(filtered, d)
	}
	dims = filtered

	sort.Slice(dims, func(i, j int) bool { return dims[i].score > dims[j].score })

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

	// Boost from currently viewing manga context
	if contextMangaID != nil {
		meta, err := s.getMangaMeta(ctx, contextMangaID.String())
		if err == nil && meta != nil {
			for _, tag := range meta.Tags {
				if _, stopped := s.stopTags[tag]; stopped {
					continue
				}
				if len(topTags) < 8 && !contains(topTags, tag) {
					topTags = append(topTags, tag)
				}
			}
			if meta.Author != "" && len(topAuthors) < 5 && !contains(topAuthors, meta.Author) {
				topAuthors = append(topAuthors, meta.Author)
			}
			if meta.Category != "" && len(topCategories) < 5 && !contains(topCategories, meta.Category) {
				topCategories = append(topCategories, meta.Category)
			}
		}
	}

	// Phase 1: fetch seen manga IDs from DB (permanent exclusion)
	var identityUUID uuid.UUID
	if id, err := uuid.Parse(identityID); err == nil {
		identityUUID = id
	}
	seenIDs, _ := s.seenRepo.ListIDsByIdentity(ctx, identityUUID)

	// Exclude context manga too
	if contextMangaID != nil {
		seenIDs = append(seenIDs, *contextMangaID)
	}

	// Phase 2+3: retrieve candidates excluding seen, ranked by popularity
	return s.querySuggestions(ctx, topTags, topAuthors, topCategories, seenIDs, limit)
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

type interestDim struct {
	key   string
	score float64
}

// loadInterests fetches interest dimensions for an identity, using Redis as cache-aside.
func (s *Store) loadInterests(ctx context.Context, identityID string) ([]interestDim, error) {
	cacheKey := interestsPrefix + identityID

	// Try Redis cache first
	raw, err := s.rdb.HGetAll(ctx, cacheKey).Result()
	if err == nil && len(raw) > 0 {
		dims := make([]interestDim, 0, len(raw))
		for k, v := range raw {
			var score float64
			fmt.Sscanf(v, "%f", &score)
			dims = append(dims, interestDim{k, score})
		}
		return dims, nil
	}

	// Cache miss — query DB
	id, err := uuid.Parse(identityID)
	if err != nil {
		return nil, nil
	}
	var rows []struct {
		Dimension string  `db:"dimension"`
		Score     float64 `db:"score"`
	}
	err = s.db.SelectContext(ctx, &rows,
		`SELECT dimension, score FROM user_interests WHERE identity_id = $1`, id)
	if err != nil || len(rows) == 0 {
		return nil, nil
	}

	// Populate cache
	args := make([]interface{}, 0, len(rows)*2)
	dims := make([]interestDim, 0, len(rows))
	for _, r := range rows {
		args = append(args, r.Dimension, fmt.Sprintf("%f", r.Score))
		dims = append(dims, interestDim{r.Dimension, r.Score})
	}
	s.rdb.HSet(ctx, cacheKey, args...)
	s.rdb.Expire(ctx, cacheKey, interestCacheTTL)

	return dims, nil
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

	// Phase 2: collect candidate IDs via separate index scans + UNION.
	// Each branch uses its own index (GIN for tags, B-tree for author/category).
	// UNION deduplicates. $1 (seen array) is reused across all three branches.
	var candidateIDs []uuid.UUID
	err := s.db.SelectContext(ctx, &candidateIDs, `
		SELECT id FROM mangas WHERE id != ALL($1::uuid[]) AND tags && $2::text[]
		UNION
		SELECT id FROM mangas WHERE id != ALL($1::uuid[]) AND author != '' AND author = ANY($3::text[])
		UNION
		SELECT id FROM mangas WHERE id != ALL($1::uuid[]) AND category != '' AND category = ANY($4::text[])`,
		pq.Array(seenIDs),
		pq.Array(tags),
		pq.Array(authors),
		pq.Array(categories),
	)
	if err != nil || len(candidateIDs) == 0 {
		return nil, err
	}

	// Phase 3: fetch full rows for candidates, rank by popularity score.
	var mangas []*model.Manga
	err = s.db.SelectContext(ctx, &mangas, `
		SELECT m.* FROM mangas m
		LEFT JOIN manga_popularity p ON p.manga_id = m.id
		WHERE m.id = ANY($1)
		ORDER BY COALESCE(p.score, 0) DESC
		LIMIT $2`,
		pq.Array(candidateIDs),
		limit,
	)
	return mangas, err
}

// coldStartSuggestions is the fallback when no interest profile exists yet.
// If a context manga is provided, returns similar manga ranked by popularity.
// Otherwise returns the most popular manga the user hasn't seen.
func (s *Store) coldStartSuggestions(ctx context.Context, contextMangaID *uuid.UUID, seenIDs []uuid.UUID, limit int) ([]*model.Manga, error) {
	if contextMangaID != nil {
		meta, err := s.getMangaMeta(ctx, contextMangaID.String())
		if err == nil && meta != nil && (len(meta.Tags) > 0 || meta.Author != "" || meta.Category != "") {
			return s.querySuggestions(ctx, meta.Tags, []string{meta.Author}, []string{meta.Category}, seenIDs, limit)
		}
	}

	// No context — return top popular unseen manga
	var mangas []*model.Manga
	err := s.db.SelectContext(ctx, &mangas, `
		SELECT m.* FROM mangas m
		LEFT JOIN manga_popularity p ON p.manga_id = m.id
		WHERE ($1::uuid[] IS NULL OR m.id != ALL($1::uuid[]))
		ORDER BY COALESCE(p.score, 0) DESC
		LIMIT $2`,
		pq.Array(seenIDs),
		limit,
	)
	return mangas, err
}

// --- Similar ---

// GetSimilar returns manga similar to the given manga_id by matching tags,
// author, or category. The source manga is excluded from the results.
func (s *Store) GetSimilar(ctx context.Context, mangaID string, limit int) ([]*model.Manga, error) {
	meta, err := s.getMangaMeta(ctx, mangaID)
	if err != nil || meta == nil {
		return nil, err
	}
	if len(meta.Tags) == 0 && meta.Author == "" && meta.Category == "" {
		return nil, nil
	}

	srcID, err := uuid.Parse(mangaID)
	if err != nil {
		return nil, err
	}

	var mangas []*model.Manga
	err = s.db.SelectContext(ctx, &mangas, `
		SELECT m.* FROM mangas m
		LEFT JOIN manga_popularity p ON p.manga_id = m.id
		WHERE m.id != $1
		  AND (
		        ($2::text[] != '{}' AND m.tags && $2::text[])
		     OR ($3 != '' AND m.author = $3)
		     OR ($4 != '' AND m.category = $4)
		  )
		ORDER BY COALESCE(p.score, 0) DESC
		LIMIT $5`,
		srcID,
		pq.Array(meta.Tags),
		meta.Author,
		meta.Category,
		limit,
	)
	return mangas, err
}

// InvalidateInterestCache removes the Redis interest cache for an identity,
// forcing the next suggestion request to reload from DB.
func (s *Store) InvalidateInterestCache(ctx context.Context, identityID uuid.UUID) {
	s.rdb.Del(ctx, interestsPrefix+identityID.String())
}

// --- Decay loop ---

// StartDecay runs the hourly trending decay in the background.
// Call it as a goroutine from serve/server.go.
func (s *Store) StartDecay(ctx context.Context) {
	ticker := time.NewTicker(s.decayInterval)
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
