package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/middleware"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/service"
)

type PageHandler struct {
	pageSvc *service.PageService
}

func NewPageHandler(pageSvc *service.PageService) *PageHandler {
	return &PageHandler{pageSvc: pageSvc}
}

// Upload godoc
//
//	@Summary	Upload individual pages
//	@Tags		page
//	@Accept		mpfd
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID		path		string	true	"Manga ID"
//	@Param		chapterID	path		string	true	"Chapter ID"
//	@Param		pages		formData	file	true	"Image files (jpeg/png/webp), field name: pages"
//	@Success	201			{array}		dto.PageUploadResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID}/pages [post]
func (h *PageHandler) Upload(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, chapterID, ok := parseMangaChapter(c)
	if !ok {
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multipart form required"})
		return
	}
	fileHeaders := form.File["pages"]
	if len(fileHeaders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one page file required"})
		return
	}

	var uploads []service.UploadFile
	for _, fh := range fileHeaders {
		f, mime, size, err := openUpload(fh)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := validateImageMIME(mime); err != nil {
			f.Close()
			respondError(c, apperror.ErrInvalidMIME)
			return
		}
		uploads = append(uploads, service.UploadFile{
			Header:  fh,
			Content: f,
			MIME:    mime,
			Size:    size,
		})
	}
	defer func() {
		for _, u := range uploads {
			if rc, ok := u.Content.(interface{ Close() error }); ok {
				rc.Close()
			}
		}
	}()

	pages, err := h.pageSvc.UploadPages(c.Request.Context(), userID, mangaID, chapterID, uploads)
	if err != nil {
		respondError(c, err)
		return
	}
	respondCreated(c, toPageUploadResponseList(pages))
}

// UploadZip godoc
//
//	@Summary	Upload pages from ZIP
//	@Description	Replaces all pages in the chapter. Files inside the ZIP are sorted by filename. An optional metadata.json at the ZIP root is parsed and returned as suggestions.
//	@Tags		page
//	@Accept		mpfd
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID		path		string	true	"Manga ID"
//	@Param		chapterID	path		string	true	"Chapter ID"
//	@Param		file		formData	file	true	"ZIP archive"
//	@Success	201			{object}	dto.ZipUploadResponse
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID}/pages/zip [post]
func (h *PageHandler) UploadZip(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, chapterID, ok := parseMangaChapter(c)
	if !ok {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "zip file is required (field: file)"})
		return
	}
	if fileHeader.Header.Get("Content-Type") != "application/zip" &&
		fileHeader.Header.Get("Content-Type") != "application/x-zip-compressed" {
		// Allow content-type to be missing or wrong — we'll try to parse anyway
		// Only hard-reject if the client sends a clearly wrong type
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open uploaded file"})
		return
	}
	defer f.Close()

	pages, meta, err := h.pageSvc.UploadZip(c.Request.Context(), userID, mangaID, chapterID, f, fileHeader.Size)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := dto.ZipUploadResponse{
		Pages: toPageUploadResponseList(pages),
	}
	if meta != nil {
		resp.MetadataSuggestions = &dto.ZipMetadataSuggestions{
			ChapterNumber: meta.ChapterNumber,
			ChapterTitle:  meta.ChapterTitle,
			Author:        meta.Author,
			Artist:        meta.Artist,
			Tags:          meta.Tags,
			Category:      meta.Category,
			Language:      meta.Language,
		}
	}
	respondCreated(c, resp)
}

// UploadOneshotZip godoc
//
//	@Summary	Upload oneshot ZIP
//	@Description	For oneshot manga only. Auto-creates the chapter (number=0), uploads pages, sets the first page as cover if none exists, and returns metadata suggestions from an optional metadata.json in the ZIP.
//	@Tags		page
//	@Accept		mpfd
//	@Produce	json
//	@Security	BearerAuth
//	@Param		mangaID	path		string	true	"Manga ID (must be type=oneshot)"
//	@Param		file	formData	file	true	"ZIP archive"
//	@Success	201		{object}	dto.OneshotUploadResponse
//	@Failure	400		{object}	dto.ErrorResponse	"Not a oneshot or invalid ZIP"
//	@Failure	401		{object}	dto.ErrorResponse
//	@Failure	403		{object}	dto.ErrorResponse
//	@Failure	409		{object}	dto.ErrorResponse	"Chapter already exists"
//	@Router		/mangas/{mangaID}/oneshot/upload [post]
func (h *PageHandler) UploadOneshotZip(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, err := uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "zip file is required (field: file)"})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open uploaded file"})
		return
	}
	defer f.Close()

	result, err := h.pageSvc.UploadOneshotZip(c.Request.Context(), userID, mangaID, f, fileHeader.Size)
	if err != nil {
		respondError(c, err)
		return
	}

	resp := dto.OneshotUploadResponse{
		Chapter: dto.NewChapterResponse(result.Chapter),
		Pages:   toPageUploadResponseList(result.Pages),
	}
	if result.Meta != nil {
		resp.MetadataSuggestions = &dto.ZipMetadataSuggestions{
			ChapterNumber: result.Meta.ChapterNumber,
			ChapterTitle:  result.Meta.ChapterTitle,
			Author:        result.Meta.Author,
			Artist:        result.Meta.Artist,
			Tags:          result.Meta.Tags,
			Category:      result.Meta.Category,
			Language:      result.Meta.Language,
		}
	}
	respondCreated(c, resp)
}

// Delete godoc
//
//	@Summary	Delete a page
//	@Tags		page
//	@Security	BearerAuth
//	@Param		mangaID		path	string	true	"Manga ID"
//	@Param		chapterID	path	string	true	"Chapter ID"
//	@Param		pageNumber	path	int		true	"Page number"
//	@Success	204			"No Content"
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Failure	404			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID}/pages/{pageNumber} [delete]
func (h *PageHandler) Delete(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, chapterID, ok := parseMangaChapter(c)
	if !ok {
		return
	}
	pageNumber, err := strconv.Atoi(c.Param("pageNumber"))
	if err != nil || pageNumber < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	if err := h.pageSvc.DeletePage(c.Request.Context(), userID, mangaID, chapterID, pageNumber); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Reorder godoc
//
//	@Summary	Reorder pages
//	@Tags		page
//	@Accept		json
//	@Security	BearerAuth
//	@Param		mangaID		path	string					true	"Manga ID"
//	@Param		chapterID	path	string					true	"Chapter ID"
//	@Param		body		body	dto.ReorderPagesRequest	true	"Ordered list of page IDs"
//	@Success	204			"No Content"
//	@Failure	400			{object}	dto.ErrorResponse
//	@Failure	401			{object}	dto.ErrorResponse
//	@Failure	403			{object}	dto.ErrorResponse
//	@Router		/mangas/{mangaID}/chapters/{chapterID}/pages/reorder [patch]
func (h *PageHandler) Reorder(c *gin.Context) {
	userID := middleware.MustUserID(c)
	mangaID, chapterID, ok := parseMangaChapter(c)
	if !ok {
		return
	}

	var req dto.ReorderPagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.pageSvc.ReorderPages(c.Request.Context(), userID, mangaID, chapterID, req.PageIDs); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// parseMangaChapter extracts and validates :mangaID and :chapterID path params.
func parseMangaChapter(c *gin.Context) (mangaID, chapterID uuid.UUID, ok bool) {
	var err error
	mangaID, err = uuid.Parse(c.Param("mangaID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manga id"})
		return uuid.Nil, uuid.Nil, false
	}
	chapterID, err = uuid.Parse(c.Param("chapterID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter id"})
		return uuid.Nil, uuid.Nil, false
	}
	return mangaID, chapterID, true
}

func toPageUploadResponseList(pages []*model.Page) []dto.PageUploadResponse {
	out := make([]dto.PageUploadResponse, len(pages))
	for i, p := range pages {
		out[i] = dto.PageUploadResponse{ID: p.ID, Number: p.Number, ObjectKey: p.ObjectKey}
	}
	return out
}
