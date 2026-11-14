package handler

import (
	"net/http"
	"strconv"

	"proem-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	articleService service.ArticleService
}

func NewAdminHandler(as service.ArticleService) *AdminHandler {
	return &AdminHandler{articleService: as}
}

type moderationReasonDTO struct {
	Reason *string `json:"reason"`
}

// GET /api/v1/admin/articles/requests
func (h *AdminHandler) ListModerationRequests(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	articles, err := h.articleService.ListModerationRequests(c.Request.Context(), limit, offset)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, articles)
}

// POST /api/v1/admin/articles/:id/approve
func (h *AdminHandler) ApproveArticle(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	if err := h.articleService.ApprovePublic(c.Request.Context(), articleID); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Article approved for public listing"})
}

// POST /api/v1/admin/articles/:id/reject
func (h *AdminHandler) RejectArticle(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var dto moderationReasonDTO
	_ = c.ShouldBindJSON(&dto) // reason опционален

	if err := h.articleService.RejectPublic(c.Request.Context(), articleID, dto.Reason); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Public publication request rejected; article remains accessible by link"})
}

// POST /api/v1/admin/articles/:id/unpublish
func (h *AdminHandler) UnpublishArticle(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var dto moderationReasonDTO
	_ = c.ShouldBindJSON(&dto) // reason опционален

	if err := h.articleService.Unpublish(c.Request.Context(), articleID, dto.Reason); err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Article unpublished and no longer accessible"})
}
