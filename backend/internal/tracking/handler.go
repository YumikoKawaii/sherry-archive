package tracking

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/pkg/token"
)

type Handler struct {
	store    Store
	tokenMgr *token.Manager
	enricher Enricher // optional, nil = disabled
}

func NewHandler(store Store, tokenMgr *token.Manager, enricher Enricher) *Handler {
	return &Handler{store: store, tokenMgr: tokenMgr, enricher: enricher}
}

// Mount registers the tracking endpoint on the given engine independently
// of the main API router.
func (h *Handler) Mount(r *gin.Engine) {
	r.POST("/api/track", h.Ingest)
}

// Ingest handles POST /api/track.
// No authentication required — user_id is extracted from JWT if present.
func (h *Handler) Ingest(c *gin.Context) {
	var req IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := h.optionalUserID(c)
	ipHash := HashIP(realIP(c.Request))
	ua := c.Request.UserAgent()
	now := time.Now()

	rows := make([]EventRow, 0, len(req.Events))
	for _, e := range req.Events {
		deviceID, err := uuid.Parse(e.DeviceID)
		if err != nil {
			continue // skip malformed device ids silently
		}
		props, _ := json.Marshal(e.Properties)
		rows = append(rows, EventRow{
			DeviceID:   deviceID,
			UserID:     userID,
			Event:      e.Event,
			Properties: props,
			Referrer:   e.Referrer,
			IPHash:     ipHash,
			UserAgent:  ua,
			CreatedAt:  now,
		})
	}

	// Fire-and-forget with a fresh context — request context is cancelled
	// the moment this handler returns, which would abort the insert.
	enricher := h.enricher
	go func() {
		ctx := context.Background()
		_ = h.store.Insert(ctx, rows)
		if enricher != nil {
			enricher.ProcessEvents(ctx, rows)
		}
	}()

	c.Status(http.StatusNoContent)
}

// optionalUserID extracts user_id from the Bearer token if valid; returns nil otherwise.
func (h *Handler) optionalUserID(c *gin.Context) *uuid.UUID {
	raw := c.GetHeader("Authorization")
	if !strings.HasPrefix(raw, "Bearer ") {
		return nil
	}
	claims, err := h.tokenMgr.ParseAccessToken(strings.TrimPrefix(raw, "Bearer "))
	if err != nil {
		return nil
	}
	id := claims.UserID
	return &id
}

// realIP resolves the client IP, respecting common proxy headers.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip, _, err := net.SplitHostPort(strings.Split(xff, ",")[0]); err == nil {
			return ip
		}
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
