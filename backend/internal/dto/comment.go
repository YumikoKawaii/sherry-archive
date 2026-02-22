package dto

import (
	"time"

	"github.com/yumikokawaii/sherry-archive/internal/model"
)

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

type CommentAuthor struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type CommentResponse struct {
	ID        string        `json:"id"`
	Content   string        `json:"content"`
	Author    CommentAuthor `json:"author"`
	Edited    bool          `json:"edited"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func NewCommentResponse(c *model.CommentWithAuthor) CommentResponse {
	return CommentResponse{
		ID:      c.ID.String(),
		Content: c.Content,
		Author: CommentAuthor{
			ID:        c.UserID.String(),
			Username:  c.AuthorUsername,
			AvatarURL: c.AuthorAvatarURL,
		},
		Edited:    c.Edited,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func NewCommentResponses(rows []*model.CommentWithAuthor) []CommentResponse {
	out := make([]CommentResponse, len(rows))
	for i, r := range rows {
		out[i] = NewCommentResponse(r)
	}
	return out
}
