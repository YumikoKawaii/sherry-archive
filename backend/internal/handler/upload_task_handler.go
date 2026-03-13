package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/dto"
	"github.com/yumikokawaii/sherry-archive/internal/service"
)

type UploadTaskHandler struct {
	uploadTaskSvc *service.UploadTaskService
}

func NewUploadTaskHandler(uploadTaskSvc *service.UploadTaskService) *UploadTaskHandler {
	return &UploadTaskHandler{uploadTaskSvc: uploadTaskSvc}
}

// GetTask godoc
//
//	@Summary	Get upload task status
//	@Tags		upload
//	@Security	BearerAuth
//	@Param		taskID	path		string	true	"Task ID"
//	@Success	200		{object}	dto.UploadTaskResponse
//	@Failure	404		{object}	dto.ErrorResponse
//	@Router		/tasks/{taskID} [get]
func (h *UploadTaskHandler) GetTask(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("taskID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	task, err := h.uploadTaskSvc.GetTask(c.Request.Context(), taskID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dto.NewUploadTaskResponse(task)})
}
