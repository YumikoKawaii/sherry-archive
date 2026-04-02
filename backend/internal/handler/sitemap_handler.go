package handler

import (
	"encoding/xml"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
)

const siteBase = "https://sherry-archive.com"

type SitemapHandler struct {
	mangaRepo   repository.MangaRepository
	chapterRepo repository.ChapterRepository
}

func NewSitemapHandler(mangaRepo repository.MangaRepository, chapterRepo repository.ChapterRepository) *SitemapHandler {
	return &SitemapHandler{mangaRepo: mangaRepo, chapterRepo: chapterRepo}
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

func (h *SitemapHandler) Sitemap(c *gin.Context) {
	ctx := c.Request.Context()

	mangas, err := h.mangaRepo.ListAllForSitemap(ctx)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	chapters, err := h.chapterRepo.ListAllForSitemap(ctx)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	urls := make([]sitemapURL, 0, 1+len(mangas)+len(chapters))
	urls = append(urls, sitemapURL{
		Loc:        siteBase + "/",
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	for _, m := range mangas {
		urls = append(urls, sitemapURL{
			Loc:        siteBase + "/manga/" + m.ID.String(),
			LastMod:    m.UpdatedAt.UTC().Format("2006-01-02"),
			ChangeFreq: "weekly",
			Priority:   "0.8",
		})
	}

	for _, ch := range chapters {
		urls = append(urls, sitemapURL{
			Loc:        siteBase + "/manga/" + ch.MangaID.String() + "/chapter/" + ch.ID.String(),
			LastMod:    ch.UpdatedAt.UTC().Format("2006-01-02"),
			ChangeFreq: "monthly",
			Priority:   "0.6",
		})
	}

	urlset := sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	out, err := xml.MarshalIndent(urlset, "", "  ")
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Data(http.StatusOK, "application/xml; charset=utf-8", append([]byte(xml.Header), out...))
}
