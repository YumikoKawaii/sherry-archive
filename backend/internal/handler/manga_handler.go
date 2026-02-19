package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/middleware"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/internal/service"
	"github.com/yumikokawaii/sherry-archive/pkg/pagination"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
)

type MangaHandler struct {
	mangaSvc *service.MangaService
	storage  *storage.Client
}

func NewMangaHandler(mangaSvc *service.MangaService, storage *storage.Client) *MangaHandler {
	return &MangaHandler{mangaSvc: mangaSvc, storage: storage}
}

// resolveCoverURL converts a stored object key to a presigned URL.
// This is a local HMAC computation in the MinIO client — no network call.
// Returns empty string if key is empty or signing fails.
func (h *MangaHandler) resolveCoverURL(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	u, err := h.storage.PresignedGetURL(ctx, key)
	if err != nil {
		return ""
	}
	return u.String()
}

func (h *MangaHandler) toResponse(ctx context.Context, m *model.Manga) dto.MangaResponse {
	return dto.NewMangaResponse(m, h.resolveCoverURL(ctx, m.CoverKey))
}

func (h *MangaHandler) toResponseList(ctx context.Context, ms []*model.Manga) []dto.MangaResponse {
	out := make([]dto.MangaResponse, len(ms))
	for i, m := range ms {
		out[i] = h.toResponse(ctx, m)
	}
	return out
}

func (h *MangaHandler) List(c *gin.Context) {
	p := pagination.FromQuery(c)
	filter := repository.MangaFilter{
		Query:    c.Query("q"),
		Status:   c.Query("status"),
		Tags:     c.QueryArray("tags[]"),
		Sort:     c.Query("sort"),
		Author:   c.Query("author"),
		Artist:   c.Query("artist"),
		Category: c.Query("category"),
	}
	mangas, total, err := h.mangaSvc.List(c.Request.Context(), filter, p)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.PagedResponse[dto.MangaResponse]{
		Items: h.toResponseList(c.Request.Context(), mangas),
		Total: total,
		Page:  p.Page,
		Limit: p.Limit,
	})
}

func (h *MangaHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	m, err := h.mangaSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, h.toResponse(c.Request.Context(), m))
}

func (h *MangaHandler) Create(c *gin.Context) {
	userID := middleware.MustUserID(c)

	var req dto.CreateMangaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Status == "" {
		req.Status = model.StatusOngoing
	}
	if req.Type == "" {
		req.Type = model.TypeSeries
	}

	m, err := h.mangaSvc.Create(c.Request.Context(), service.CreateMangaInput{
		OwnerID:     userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Type:        req.Type,
		Tags:        req.Tags,
		Author:      req.Author,
		Artist:      req.Artist,
		Category:    req.Category,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	// Newly created manga has no cover yet — cover_url will be ""
	respondCreated(c, h.toResponse(c.Request.Context(), m))
}

func (h *MangaHandler) Update(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	var req dto.UpdateMangaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	m, err := h.mangaSvc.Update(c.Request.Context(), userID, mangaID, service.UpdateMangaInput{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Type:        req.Type,
		Tags:        req.Tags,
		Author:      req.Author,
		Artist:      req.Artist,
		Category:    req.Category,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, h.toResponse(c.Request.Context(), m))
}

func (h *MangaHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	if err := h.mangaSvc.Delete(c.Request.Context(), userID, mangaID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MangaHandler) UpdateCover(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	fileHeader, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover file is required"})
		return
	}

	f, mime, size, err := openUpload(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	if err := validateImageMIME(mime); err != nil {
		respondError(c, err)
		return
	}

	// Store by object key — presigned URL is resolved at read time
	objectKey := fmt.Sprintf("covers/%s/%s", mangaID, uuid.Must(uuid.NewV7()))
	if err := h.storage.PutObject(c.Request.Context(), objectKey, mime, f, size); err != nil {
		respondError(c, err)
		return
	}

	m, err := h.mangaSvc.UpdateCover(c.Request.Context(), userID, mangaID, objectKey)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, h.toResponse(c.Request.Context(), m))
}

func (h *MangaHandler) ListByUser(c *gin.Context) {
	ownerID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	p := pagination.FromQuery(c)
	mangas, total, err := h.mangaSvc.ListByOwner(c.Request.Context(), ownerID, p)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.PagedResponse[dto.MangaResponse]{
		Items: h.toResponseList(c.Request.Context(), mangas),
		Total: total,
		Page:  p.Page,
		Limit: p.Limit,
	})
}
