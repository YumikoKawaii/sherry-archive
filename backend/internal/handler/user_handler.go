package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/middleware"
	"github.com/yumikokawaii/sherry-archive/internal/service"
	"github.com/yumikokawaii/sherry-archive/pkg/storage"
)

type UserHandler struct {
	userSvc *service.UserService
	storage *storage.Client
}

func NewUserHandler(userSvc *service.UserService, storage *storage.Client) *UserHandler {
	return &UserHandler{userSvc: userSvc, storage: storage}
}

// GetUser godoc
//
//	@Summary	Get public user profile
//	@Tags		user
//	@Produce	json
//	@Param		userID	path		string	true	"User ID"
//	@Success	200		{object}	dto.PublicUserResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/users/{userID} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := h.userSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewPublicUserResponse(user))
}

// UpdateMe godoc
//
//	@Summary	Update own profile
//	@Tags		user
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		body	body		dto.UpdateUserRequest	true	"Profile fields to update"
//	@Success	200		{object}	dto.UserResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Router		/users/me [patch]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := middleware.MustUserID(c)

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userSvc.Update(c.Request.Context(), userID, service.UpdateUserInput{
		Username: req.Username,
		Bio:      req.Bio,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewUserResponse(user))
}

// UpdateAvatar godoc
//
//	@Summary	Upload avatar
//	@Tags		user
//	@Accept		mpfd
//	@Produce	json
//	@Security	BearerAuth
//	@Param		avatar	formData	file	true	"Avatar image (jpeg/png/webp)"
//	@Success	200		{object}	dto.UserResponse
//	@Failure	400		{object}	dto.ErrorResponse
//	@Failure	401		{object}	dto.ErrorResponse
//	@Router		/users/me/avatar [put]
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID := middleware.MustUserID(c)

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar file is required"})
		return
	}

	f, mime, size, err := openUpload(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	if err := validateImageMIME(mime); err != nil {
		respondError(c, apperror.ErrInvalidMIME)
		return
	}

	objectKey := fmt.Sprintf("avatars/%s/%s", userID, uuid.Must(uuid.NewV7()))
	if err := h.storage.PutObject(c.Request.Context(), objectKey, mime, f, size); err != nil {
		respondError(c, err)
		return
	}

	u, err := h.storage.PresignedGetURL(c.Request.Context(), objectKey)
	if err != nil {
		respondError(c, err)
		return
	}

	user, err := h.userSvc.UpdateAvatar(c.Request.Context(), userID, u.String())
	if err != nil {
		respondError(c, err)
		return
	}
	respondOK(c, dto.NewUserResponse(user))
}
