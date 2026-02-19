package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
	"golang.org/x/sync/errgroup"
)

type OneshotUploadResult struct {
	Chapter *model.Chapter
	Pages   []*model.Page
	Meta    *ZipMetadata
}

var allowedExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

type PageService struct {
	pageRepo    repository.PageRepository
	chapterRepo repository.ChapterRepository
	mangaRepo   repository.MangaRepository
	storage     *storage.Client
}

func NewPageService(
	pageRepo repository.PageRepository,
	chapterRepo repository.ChapterRepository,
	mangaRepo repository.MangaRepository,
	storage *storage.Client,
) *PageService {
	return &PageService{
		pageRepo:    pageRepo,
		chapterRepo: chapterRepo,
		mangaRepo:   mangaRepo,
		storage:     storage,
	}
}

type UploadFile struct {
	Header  *multipart.FileHeader
	Content io.Reader
	MIME    string
	Size    int64
}

// UploadPages appends individual files to a chapter, numbered after any existing pages.
func (s *PageService) UploadPages(ctx context.Context, requesterID, mangaID, chapterID uuid.UUID, files []UploadFile) ([]*model.Page, error) {
	if err := s.checkOwnership(ctx, requesterID, mangaID, chapterID); err != nil {
		return nil, err
	}

	existing, err := s.pageRepo.CountByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	pages := make([]*model.Page, len(files))
	for i := range files {
		pages[i] = &model.Page{
			ID:        uuid.Must(uuid.NewV7()),
			ChapterID: chapterID,
			Number:    existing + i + 1,
			ObjectKey: pageObjectKey(mangaID, chapterID),
			CreatedAt: time.Now(),
		}
	}

	return s.uploadAndPersist(ctx, chapterID, pages, files)
}

// UploadZip replaces all pages of a chapter from a zip archive.
// Files inside the zip are sorted by filename to determine page order,
// so naming them 001.jpg, 002.jpg, … gives a deterministic result.
// An optional metadata.json at the ZIP root is parsed and returned as suggestions.
func (s *PageService) UploadZip(ctx context.Context, requesterID, mangaID, chapterID uuid.UUID, r io.ReaderAt, size int64) ([]*model.Page, *ZipMetadata, error) {
	if err := s.checkOwnership(ctx, requesterID, mangaID, chapterID); err != nil {
		return nil, nil, err
	}

	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, nil, apperror.ErrBadRequest
	}

	// Extract optional metadata (best-effort, errors ignored)
	meta, _ := extractZipMetadata(r, size)

	// Collect valid image entries, sort by filename
	type zipEntry struct {
		f    *zip.File
		mime string
	}
	var entries []zipEntry
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		mime, ok := allowedExtensions[ext]
		if !ok {
			continue
		}
		entries = append(entries, zipEntry{f: f, mime: mime})
	}
	if len(entries) == 0 {
		return nil, nil, apperror.ErrBadRequest
	}
	sort.Slice(entries, func(i, j int) bool {
		return filepath.Base(entries[i].f.Name) < filepath.Base(entries[j].f.Name)
	})

	// Delete existing pages first (replace semantics)
	existing, err := s.pageRepo.GetByChapter(ctx, chapterID)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range existing {
		_ = s.storage.DeleteObject(ctx, p.ObjectKey) // best-effort
		_ = s.pageRepo.Delete(ctx, p.ID)
	}

	// Build page records and upload in parallel
	pages := make([]*model.Page, len(entries))
	for i := range entries {
		pages[i] = &model.Page{
			ID:        uuid.Must(uuid.NewV7()),
			ChapterID: chapterID,
			Number:    i + 1,
			ObjectKey: pageObjectKey(mangaID, chapterID),
			CreatedAt: time.Now(),
		}
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for i, e := range entries {
		i, e := i, e
		eg.Go(func() error {
			rc, err := e.f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			return s.storage.PutObject(egCtx, pages[i].ObjectKey, e.mime, rc, int64(e.f.UncompressedSize64))
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	if err := s.pageRepo.CreateBatch(ctx, pages); err != nil {
		return nil, nil, err
	}
	if err := s.chapterRepo.UpdatePageCount(ctx, chapterID, len(pages)); err != nil {
		return nil, nil, err
	}

	return pages, meta, nil
}

// UploadOneshotZip creates the oneshot chapter (if not yet existing) and uploads
// pages from the ZIP in a single operation. The chapter title is taken from
// metadata.json inside the ZIP when present.
func (s *PageService) UploadOneshotZip(ctx context.Context, requesterID, mangaID uuid.UUID, r io.ReaderAt, size int64) (*OneshotUploadResult, error) {
	manga, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	if manga.OwnerID != requesterID {
		return nil, apperror.ErrForbidden
	}
	if manga.Type != model.TypeOneshot {
		return nil, apperror.ErrBadRequest
	}

	existing, err := s.chapterRepo.ListByManga(ctx, mangaID)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return nil, apperror.ErrConflict
	}

	meta, _ := extractZipMetadata(r, size)

	chapterTitle := "Oneshot"
	if meta != nil && meta.ChapterTitle != "" {
		chapterTitle = meta.ChapterTitle
	}

	now := time.Now()
	ch := &model.Chapter{
		ID:        uuid.Must(uuid.NewV7()),
		MangaID:   mangaID,
		Number:    0,
		Title:     chapterTitle,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.chapterRepo.Create(ctx, ch); err != nil {
		return nil, err
	}

	pages, _, err := s.UploadZip(ctx, requesterID, mangaID, ch.ID, r, size)
	if err != nil {
		return nil, err
	}

	// Auto-set cover from the first page (best-effort — don't fail the upload if this goes wrong)
	if len(pages) > 0 && manga.CoverKey == "" {
		manga.CoverKey = pages[0].ObjectKey
		manga.UpdatedAt = time.Now()
		_ = s.mangaRepo.Update(ctx, manga)
	}

	return &OneshotUploadResult{Chapter: ch, Pages: pages, Meta: meta}, nil
}

func (s *PageService) DeletePage(ctx context.Context, requesterID, mangaID, chapterID uuid.UUID, pageNumber int) error {
	if err := s.checkOwnership(ctx, requesterID, mangaID, chapterID); err != nil {
		return err
	}

	page, err := s.pageRepo.GetByChapterAndNumber(ctx, chapterID, pageNumber)
	if err != nil {
		return err
	}

	if err := s.storage.DeleteObject(ctx, page.ObjectKey); err != nil {
		return err
	}
	if err := s.pageRepo.Delete(ctx, page.ID); err != nil {
		return err
	}

	count, err := s.pageRepo.CountByChapter(ctx, chapterID)
	if err != nil {
		return err
	}
	return s.chapterRepo.UpdatePageCount(ctx, chapterID, count)
}

func (s *PageService) ReorderPages(ctx context.Context, requesterID, mangaID, chapterID uuid.UUID, pageIDs []uuid.UUID) error {
	if err := s.checkOwnership(ctx, requesterID, mangaID, chapterID); err != nil {
		return err
	}
	return s.pageRepo.UpdateNumbers(ctx, chapterID, pageIDs)
}

func (s *PageService) GetPagesWithURLs(ctx context.Context, chapterID uuid.UUID) ([]*model.Page, []string, error) {
	pages, err := s.pageRepo.GetByChapter(ctx, chapterID)
	if err != nil {
		return nil, nil, err
	}

	urls := make([]string, len(pages))
	for i, p := range pages {
		u, err := s.storage.PresignedGetURL(ctx, p.ObjectKey)
		if err != nil {
			return nil, nil, err
		}
		urls[i] = u.String()
	}
	return pages, urls, nil
}

// checkOwnership verifies the requester owns the manga and the chapter belongs to it.
func (s *PageService) checkOwnership(ctx context.Context, requesterID, mangaID, chapterID uuid.UUID) error {
	manga, err := s.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return err
	}
	if manga.OwnerID != requesterID {
		return apperror.ErrForbidden
	}
	ch, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return err
	}
	if ch.MangaID != mangaID {
		return apperror.ErrNotFound
	}
	return nil
}

func (s *PageService) uploadAndPersist(ctx context.Context, chapterID uuid.UUID, pages []*model.Page, files []UploadFile) ([]*model.Page, error) {
	eg, egCtx := errgroup.WithContext(ctx)
	for i, f := range files {
		i, f := i, f
		eg.Go(func() error {
			return s.storage.PutObject(egCtx, pages[i].ObjectKey, f.MIME, f.Content, f.Size)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if err := s.pageRepo.CreateBatch(ctx, pages); err != nil {
		return nil, err
	}

	count, err := s.pageRepo.CountByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}
	if err := s.chapterRepo.UpdatePageCount(ctx, chapterID, count); err != nil {
		return nil, err
	}

	return pages, nil
}

func pageObjectKey(mangaID, chapterID uuid.UUID) string {
	return fmt.Sprintf("mangas/%s/chapters/%s/%s.webp", mangaID, chapterID, uuid.Must(uuid.NewV7()))
}
