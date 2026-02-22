package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yumikokawaii/sherry-archive/internal/middleware"
	"github.com/yumikokawaii/sherry-archive/pkg/token"
)

type Handlers struct {
	Auth     *AuthHandler
	Manga    *MangaHandler
	Chapter  *ChapterHandler
	Page     *PageHandler
	Bookmark *BookmarkHandler
	User     *UserHandler
	Comment  *CommentHandler
}

func SetupRouter(h Handlers, tokenMgr *token.Manager) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authMW := middleware.Auth(tokenMgr)

	v1 := r.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
		auth.POST("/logout", authMW, h.Auth.Logout)
		auth.GET("/me", authMW, h.Auth.Me)
	}

	// Manga routes
	mangas := v1.Group("/mangas")
	{
		mangas.GET("", h.Manga.List)
		mangas.POST("", authMW, h.Manga.Create)
		mangas.GET("/:mangaID", h.Manga.Get)
		mangas.PATCH("/:mangaID", authMW, h.Manga.Update)
		mangas.DELETE("/:mangaID", authMW, h.Manga.Delete)
		mangas.PUT("/:mangaID/cover", authMW, h.Manga.UpdateCover)

		// Oneshot direct upload
		mangas.POST("/:mangaID/oneshot/upload", authMW, h.Page.UploadOneshotZip)

		// Chapter routes
		mangas.GET("/:mangaID/chapters", h.Chapter.List)
		mangas.POST("/:mangaID/chapters", authMW, h.Chapter.Create)
		mangas.GET("/:mangaID/chapters/:chapterID", h.Chapter.Get)
		mangas.PATCH("/:mangaID/chapters/:chapterID", authMW, h.Chapter.Update)
		mangas.DELETE("/:mangaID/chapters/:chapterID", authMW, h.Chapter.Delete)

		// Page routes
		mangas.POST("/:mangaID/chapters/:chapterID/pages", authMW, h.Page.Upload)
		mangas.POST("/:mangaID/chapters/:chapterID/pages/zip", authMW, h.Page.UploadZip)
		mangas.DELETE("/:mangaID/chapters/:chapterID/pages/:pageNumber", authMW, h.Page.Delete)
		mangas.PATCH("/:mangaID/chapters/:chapterID/pages/reorder", authMW, h.Page.Reorder)

		// Comment routes
		mangas.GET("/:mangaID/comments", h.Comment.ListManga)
		mangas.POST("/:mangaID/comments", authMW, h.Comment.CreateManga)
		mangas.PATCH("/:mangaID/comments/:commentID", authMW, h.Comment.Update)
		mangas.DELETE("/:mangaID/comments/:commentID", authMW, h.Comment.Delete)

		mangas.GET("/:mangaID/chapters/:chapterID/comments", h.Comment.ListChapter)
		mangas.POST("/:mangaID/chapters/:chapterID/comments", authMW, h.Comment.CreateChapter)
	}

	// User routes
	users := v1.Group("/users")
	{
		users.GET("/:userID", h.User.GetUser)
		users.GET("/:userID/mangas", h.Manga.ListByUser)
		users.PATCH("/me", authMW, h.User.UpdateMe)
		users.PUT("/me/avatar", authMW, h.User.UpdateAvatar)
	}

	// Bookmark routes
	bookmarks := v1.Group("/users/me/bookmarks", authMW)
	{
		bookmarks.GET("", h.Bookmark.List)
		bookmarks.GET("/:mangaID", h.Bookmark.Get)
		bookmarks.PUT("/:mangaID", h.Bookmark.Upsert)
		bookmarks.DELETE("/:mangaID", h.Bookmark.Delete)
	}

	return r
}
