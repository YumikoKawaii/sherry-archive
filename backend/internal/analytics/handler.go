package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
)

type Handler struct {
	store   *Store
	storage *storage.Client
}

func NewHandler(store *Store, storage *storage.Client) *Handler {
	return &Handler{store: store, storage: storage}
}

func (h *Handler) Mount(r *gin.Engine) {
	g := r.Group("/api/v1/analytics")
	g.GET("/trending", h.Trending)
	g.GET("/suggestions", h.Suggestions)
	g.GET("/similar", h.Similar)
}

// Trending returns the top N manga ranked by recent activity score.
func (h *Handler) Trending(c *gin.Context) {
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

	out := make([]trendingItem, 0, len(results))
	for _, r := range results {
		coverURL := ""
		if r.Manga.CoverKey != "" {
			if u, err := h.storage.PresignedGetURL(c.Request.Context(), r.Manga.CoverKey); err == nil {
				coverURL = u.String()
			}
		}
		out = append(out, trendingItem{
			MangaResponse: dto.NewMangaResponse(r.Manga, coverURL),
			TrendingScore: r.Score,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// Suggestions returns personalised manga for a given device_id.
func (h *Handler) Suggestions(c *gin.Context) {
	deviceID := c.Query("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}
	limit := parseLimit(c, 12)

	mangas, err := h.store.GetSuggestions(c.Request.Context(), deviceID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]dto.MangaResponse, 0, len(mangas))
	for _, m := range mangas {
		coverURL := ""
		if m.CoverKey != "" {
			if u, err := h.storage.PresignedGetURL(c.Request.Context(), m.CoverKey); err == nil {
				coverURL = u.String()
			}
		}
		out = append(out, dto.NewMangaResponse(m, coverURL))
	}

	c.JSON(http.StatusOK, gin.H{"data": out})
}

// Similar returns manga similar to a given manga_id by shared tags, author, or category.
func (h *Handler) Similar(c *gin.Context) {
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

	out := make([]dto.MangaResponse, 0, len(mangas))
	for _, m := range mangas {
		coverURL := ""
		if m.CoverKey != "" {
			if u, err := h.storage.PresignedGetURL(c.Request.Context(), m.CoverKey); err == nil {
				coverURL = u.String()
			}
		}
		out = append(out, dto.NewMangaResponse(m, coverURL))
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
