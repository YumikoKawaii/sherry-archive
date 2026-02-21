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

// List godoc
//
//	@Summary	List chapters for a manga
//	@Tags		chapter
//	@Produce	json
//	@Param		mangaID	path		string	true	"Manga ID"
//	@Success	200		{array}		dto.ChapterResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters [get]
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

// Get godoc
//
//	@Summary	Get chapter with pages
//	@Tags		chapter
//	@Produce	json
//	@Param		mangaID		path		string	true	"Manga ID"
//	@Param		chapterID	path		string	true	"Chapter ID"
//	@Success	200			{object}	dto.ChapterWithPagesResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID} [get]
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

// Create godoc
//
//	@Summary	Create chapter
//	@Tags		chapter
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID	path		string					true	"Manga ID"
//	@Param		body	body		dto.CreateChapterRequest	true	"Chapter data (number omitted for oneshot manga)"
//	@Success	201		{object}	dto.ChapterResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	403		{object}	dto.ErrorResponse
//	@Failure	409		{object}	dto.ErrorResponse	"Duplicate chapter number or oneshot already has a chapter"
//	@Router		/mangas/{mangaID}/chapters [post]
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

// Update godoc
//
//	@Summary	Update chapter
//	@Tags		chapter
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID		path		string					true	"Manga ID"
//	@Param		chapterID	path		string					true	"Chapter ID"
//	@Param		body		body		dto.UpdateChapterRequest	true	"Fields to update"
//	@Success	200			{object}	dto.ChapterResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID} [patch]
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

// Delete godoc
//
//	@Summary	Delete chapter
//	@Tags		chapter
//	@Security	BearerAuth
//	@Param		mangaID		path	string	true	"Manga ID"
//	@Param		chapterID	path	string	true	"Chapter ID"
//	@Success	204			"No Content"
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID} [delete]
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
