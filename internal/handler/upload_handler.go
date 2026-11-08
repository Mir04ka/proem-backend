package handler

import (
	"net/http"

	"proem-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	storageService service.StorageService
}

func NewUploadHandler(ss service.StorageService) *UploadHandler {
	return &UploadHandler{storageService: ss}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	url, err := h.storageService.UploadImage(file)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}
