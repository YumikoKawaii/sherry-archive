package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/middleware"
	"github.com/yumikokawaii/sherry-archive/internal/service"
	"github.com/yumikokawaii/sherry-archive/pkg/pagination"
)

type CommentHandler struct {
	commentSvc *service.CommentService
}

func NewCommentHandler(commentSvc *service.CommentService) *CommentHandler {
	return &CommentHandler{commentSvc: commentSvc}
}

// ListMangaComments godoc
//
//	@Summary	List manga-level comments
//	@Tags		comment
//	@Produce	json
//	@Param		mangaID	path		string	true	"Manga ID"
//	@Param		page	query		int		false	"Page"
//	@Param		limit	query		int		false	"Limit"
//	@Success	200		{object}	dto.PagedResponse[dto.CommentResponse]
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/comments [get]
func (h *CommentHandler) ListManga(c *gin.Context) {
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	p := pagination.FromQuery(c)
	rows, total, err := h.commentSvc.ListByManga(c.Request.Context(), mangaID, p)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.PagedResponse[dto.CommentResponse]{
		Items: dto.NewCommentResponses(rows),
		Total: total,
		Page:  p.Page,
		Limit: p.Limit,
	})
}

// CreateMangaComment godoc
//
//	@Summary	Post a manga-level comment
//	@Tags		comment
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID	path		string						true	"Manga ID"
//	@Param		body	body		dto.CreateCommentRequest	true	"Comment"
//	@Success	201		{object}	dto.CommentResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/comments [post]
func (h *CommentHandler) CreateManga(c *gin.Context) {
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.MustUserID(c)
	comment, err := h.commentSvc.CreateMangaComment(c.Request.Context(), userID, mangaID, req.Content)
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, dto.NewCommentResponse(comment))
}

// ListChapterComments godoc
//
//	@Summary	List chapter comments
//	@Tags		comment
//	@Produce	json
//	@Param		mangaID		path		string	true	"Manga ID"
//	@Param		chapterID	path		string	true	"Chapter ID"
//	@Param		page		query		int		false	"Page"
//	@Param		limit		query		int		false	"Limit"
//	@Success	200			{object}	dto.PagedResponse[dto.CommentResponse]
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID}/comments [get]
func (h *CommentHandler) ListChapter(c *gin.Context) {
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
	p := pagination.FromQuery(c)
	rows, total, err := h.commentSvc.ListByChapter(c.Request.Context(), mangaID, chapterID, p)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.PagedResponse[dto.CommentResponse]{
		Items: dto.NewCommentResponses(rows),
		Total: total,
		Page:  p.Page,
		Limit: p.Limit,
	})
}

// CreateChapterComment godoc
//
//	@Summary	Post a chapter comment
//	@Tags		comment
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID		path		string						true	"Manga ID"
//	@Param		chapterID	path		string						true	"Chapter ID"
//	@Param		body		body		dto.CreateCommentRequest	true	"Comment"
//	@Success	201			{object}	dto.CommentResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID}/comments [post]
func (h *CommentHandler) CreateChapter(c *gin.Context) {
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
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.MustUserID(c)
	comment, err := h.commentSvc.CreateChapterComment(c.Request.Context(), userID, mangaID, chapterID, req.Content)
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, dto.NewCommentResponse(comment))
}

// UpdateComment godoc
//
//	@Summary	Edit a comment (owner only)
//	@Tags		comment
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID		path		string						true	"Manga ID"
//	@Param		commentID	path		string						true	"Comment ID"
//	@Param		body		body		dto.UpdateCommentRequest	true	"Updated content"
//	@Success	200			{object}	dto.CommentResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/comments/{commentID} [patch]
func (h *CommentHandler) Update(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("commentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}
	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := middleware.MustUserID(c)
	comment, err := h.commentSvc.Update(c.Request.Context(), userID, commentID, req.Content)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewCommentResponse(comment))
}

// DeleteComment godoc
//
//	@Summary	Delete a comment (comment owner or manga owner)
//	@Tags		comment
//	@Security	BearerAuth
//	@Param		mangaID		path	string	true	"Manga ID"
//	@Param		commentID	path	string	true	"Comment ID"
//	@Success	204
//	@Failure	401	{object}	dto.ErrorResponse
//	@Failure	403	{object}	dto.ErrorResponse
//	@Failure	404	{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/comments/{commentID} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("commentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
		return
	}
	userID := middleware.MustUserID(c)
	if err := h.commentSvc.Delete(c.Request.Context(), userID, commentID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
