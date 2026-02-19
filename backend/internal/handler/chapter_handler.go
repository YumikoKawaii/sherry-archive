package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/middleware"
	"github.com/yumikokawaii/sherry-archive/internal/service"
)

type ChapterHandler struct {
	chapterSvc *service.ChapterService
	pageSvc    *service.PageService
}

func NewChapterHandler(chapterSvc *service.ChapterService, pageSvc *service.PageService) *ChapterHandler {
	return &ChapterHandler{chapterSvc: chapterSvc, pageSvc: pageSvc}
}

func (h *ChapterHandler) List(c *gin.Context) {
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	chapters, err := h.chapterSvc.ListByManga(c.Request.Context(), mangaID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewChapterResponseList(chapters))
}

func (h *ChapterHandler) Get(c *gin.Context) {
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	chapterID, err := uuid.Parse(c.Param("chapterID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return
	}

	ch, err := h.chapterSvc.GetByID(c.Request.Context(), chapterID)
	if err != nil {
		respondError(c, err)
		return
	}
	if ch.MangaID != mangaID {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	pages, urls, err := h.pageSvc.GetPagesWithURLs(c.Request.Context(), chapterID)
	if err != nil {
		respondError(c, err)
		return
	}

	pageItems := make([]dto.PageItemResponse, len(pages))
	for i, p := range pages {
		pageItems[i] = dto.PageItemResponse{
			ID:     p.ID,
			Number: p.Number,
			URL:    urls[i],
			Width:  p.Width,
			Height: p.Height,
		}
	}

	respondOK(c, dto.ChapterWithPagesResponse{
		Chapter: dto.NewChapterResponse(ch),
		Pages:   pageItems,
	})
}

func (h *ChapterHandler) Create(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	var req dto.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.chapterSvc.Create(c.Request.Context(), userID, service.CreateChapterInput{
		MangaID: mangaID,
		Number:  req.Number,
		Title:   req.Title,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, dto.NewChapterResponse(ch))
}

func (h *ChapterHandler) Update(c *gin.Context) {
	userID := middleware.MustUserID(c)
	chapterID, err := uuid.Parse(c.Param("chapterID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return
	}

	var req dto.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ch, err := h.chapterSvc.Update(c.Request.Context(), userID, chapterID, service.UpdateChapterInput{
		Number: req.Number,
		Title:  req.Title,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewChapterResponse(ch))
}

func (h *ChapterHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	chapterID, err := uuid.Parse(c.Param("chapterID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return
	}
	if err := h.chapterSvc.Delete(c.Request.Context(), userID, chapterID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
