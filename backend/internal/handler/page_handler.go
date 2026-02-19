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

// Upload accepts multiple image files as multipart/form-data (field name: "pages").
// Pages are appended to the chapter in the order the files are received.
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

// UploadZip accepts a single ZIP archive (field name: "file") and replaces all
// pages in the chapter. Files inside the ZIP are sorted by filename to determine
// page order — name them 001.jpg, 002.jpg, … for predictable results.
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

// UploadOneshotZip accepts a ZIP archive for a oneshot manga, auto-creates the
// chapter, uploads pages, and returns chapter + metadata suggestions in one call.
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
