package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/middleware"
	"github.com/yumikokawaii/sherry-archive/internal/service"
)

type BookmarkHandler struct {
	bookmarkSvc *service.BookmarkService
}

func NewBookmarkHandler(bookmarkSvc *service.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{bookmarkSvc: bookmarkSvc}
}

// List godoc
//
//	@Summary	List bookmarks
//	@Tags		bookmark
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{array}		dto.BookmarkResponse
//	@Failure	401	{object}	dto.ErrorResponse
//	@Router		/users/me/bookmarks [get]
func (h *BookmarkHandler) List(c *gin.Context) {
	userID := middleware.MustUserID(c)
	bookmarks, err := h.bookmarkSvc.List(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewBookmarkResponseList(bookmarks))
}

// Get godoc
//
//	@Summary	Get bookmark for a manga
//	@Tags		bookmark
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID	path		string	true	"Manga ID"
//	@Success	200		{object}	dto.BookmarkResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/users/me/bookmarks/{mangaID} [get]
func (h *BookmarkHandler) Get(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	b, err := h.bookmarkSvc.Get(c.Request.Context(), userID, mangaID)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewBookmarkResponse(b))
}

// Upsert godoc
//
//	@Summary	Create or update bookmark
//	@Tags		bookmark
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID	path		string						true	"Manga ID"
//	@Param		body	body		dto.UpsertBookmarkRequest	true	"Bookmark data"
//	@Success	200		{object}	dto.BookmarkResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Router		/users/me/bookmarks/{mangaID} [put]
func (h *BookmarkHandler) Upsert(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	var req dto.UpsertBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, err := h.bookmarkSvc.Upsert(c.Request.Context(), userID, mangaID, service.UpsertBookmarkInput{
		ChapterID:      req.ChapterID,
		LastPageNumber: req.LastPageNumber,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewBookmarkResponse(b))
}

// Delete godoc
//
//	@Summary	Delete bookmark
//	@Tags		bookmark
//	@Security	BearerAuth
//	@Param		mangaID	path	string	true	"Manga ID"
//	@Success	204		"No Content"
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Router		/users/me/bookmarks/{mangaID} [delete]
func (h *BookmarkHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	if err := h.bookmarkSvc.Delete(c.Request.Context(), userID, mangaID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
