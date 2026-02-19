package service

import (
	"archive/zip"
	"encoding/json"
	"io"
	"path/filepath"
)

// ZipMetadata holds optional fields parsed from a metadata.json file at the ZIP root.
type ZipMetadata struct {
	ChapterNumber *float64 `json:"chapter_number"`
	ChapterTitle  string   `json:"chapter_title"`
	Author        string   `json:"author"`
	Artist        string   `json:"artist"`
	Tags          []string `json:"tags"`
	Category      string   `json:"category"`
	Language      string   `json:"language"`
}

// extractZipMetadata looks for a metadata.json at the ZIP root and parses it.
// Returns nil (not an error) when the file is absent.
func extractZipMetadata(r io.ReaderAt, size int64) (*ZipMetadata, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, nil
	}

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// Only consider files directly at the root (no directory component)
		if filepath.Dir(f.Name) != "." {
			continue
		}
		if filepath.Base(f.Name) != "metadata.json" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, nil
		}
		defer rc.Close()

		var meta ZipMetadata
		if err := json.NewDecoder(rc).Decode(&meta); err != nil {
			return nil, nil
		}
		return &meta, nil
	}
	return nil, nil
}
