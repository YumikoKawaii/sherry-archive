package handler

import (
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
)

func respondError(c *gin.Context, err error) {
	c.JSON(apperror.HTTPStatus(err), gin.H{"error": err.Error()})
}

func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func respondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// openUpload opens a multipart file header and returns the file, MIME type, and size.
func openUpload(fh *multipart.FileHeader) (multipart.File, string, int64, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, "", 0, err
	}
	mime := fh.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	return f, mime, fh.Size, nil
}

// validateImageMIME returns ErrInvalidMIME if the MIME type is not an accepted image format.
func validateImageMIME(mime string) error {
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	if !allowed[mime] {
		return apperror.ErrInvalidMIME
	}
	return nil
}
