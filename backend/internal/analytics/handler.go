package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/metrics"
	"github.com/yumikokawaii/sherry-archive/pkg/urlcache"
)

type Handler struct {
	store    *Store
	urlCache *urlcache.URLCache
}

func NewHandler(store *Store, urlCache *urlcache.URLCache) *Handler {
	return &Handler{store: store, urlCache: urlCache}
}

func (h *Handler) Mount(r *gin.Engine) {
	g := r.Group("/api/v1/analytics")
	g.GET("/trending", h.Trending)
	g.GET("/suggestions", h.Suggestions)
	g.GET("/similar", h.Similar)
}

// Trending returns the top N manga ranked by recent activity score.
func (h *Handler) Trending(c *gin.Context) {
	metrics.RecordAnalyticsRequest("trending")
	limit := parseLimit(c, 12)

	results, err := h.store.GetTrending(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type trendingItem struct {
		dto.MangaResponse
		TrendingScore float64 `json:"trending_score"`
	}

	keys := make([]string, len(results))
	for i, r := range results {
		keys[i] = r.Manga.CoverKey
	}
	urls, _ := h.urlCache.ResolveMany(c.Request.Context(), keys)

	out := make([]trendingItem, 0, len(results))
	for i, r := range results {
		url := ""
		if i < len(urls) {
			url = urls[i]
		}
		out = append(out, trendingItem{
			MangaResponse: dto.NewMangaResponse(r.Manga, url),
			TrendingScore: r.Score,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// Suggestions returns personalised manga for a given device_id, optionally scoped to a user_id.
func (h *Handler) Suggestions(c *gin.Context) {
	metrics.RecordAnalyticsRequest("suggestions")
	deviceID := c.Query("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}
	limit := parseLimit(c, 12)

	var userID *uuid.UUID
	if raw := c.Query("user_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			userID = &id
		}
	}

	var contextMangaID *uuid.UUID
	if raw := c.Query("manga_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			contextMangaID = &id
		}
	}

	mangas, err := h.store.GetSuggestions(c.Request.Context(), userID, deviceID, contextMangaID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	keys := make([]string, len(mangas))
	for i, m := range mangas {
		keys[i] = m.CoverKey
	}
	urls, _ := h.urlCache.ResolveMany(c.Request.Context(), keys)

	out := make([]dto.MangaResponse, 0, len(mangas))
	for i, m := range mangas {
		url := ""
		if i < len(urls) {
			url = urls[i]
		}
		out = append(out, dto.NewMangaResponse(m, url))
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// Similar returns manga similar to a given manga_id by shared tags, author, or category.
func (h *Handler) Similar(c *gin.Context) {
	metrics.RecordAnalyticsRequest("similar")
	mangaID := c.Query("manga_id")
	if mangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id is required"})
		return
	}
	limit := parseLimit(c, 8)

	mangas, err := h.store.GetSimilar(c.Request.Context(), mangaID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	keys := make([]string, len(mangas))
	for i, m := range mangas {
		keys[i] = m.CoverKey
	}
	urls, _ := h.urlCache.ResolveMany(c.Request.Context(), keys)

	out := make([]dto.MangaResponse, 0, len(mangas))
	for i, m := range mangas {
		url := ""
		if i < len(urls) {
			url = urls[i]
		}
		out = append(out, dto.NewMangaResponse(m, url))
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

func parseLimit(c *gin.Context, def int) int {
	v := c.Query("limit")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 || n > 50 {
		return def
	}
	return n
}
