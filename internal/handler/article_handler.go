package handler

import (
	"net/http"
	"strconv"

	"proem-backend/internal/middleware"
	"proem-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ArticleHandler struct {
	articleService service.ArticleService
}

func NewArticleHandler(as service.ArticleService) *ArticleHandler {
	return &ArticleHandler{articleService: as}
}

type articleUpsertDTO struct {
	Title   string `json:"title" binding:"required,min=3"`
	Content string `json:"content" binding:"required"`
}

// Publish — публикация новой статьи (только для авторизованных)
func (h *ArticleHandler) Publish(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	authorID := userIDVal.(uuid.UUID)

	var dto articleUpsertDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.articleService.Publish(c.Request.Context(), authorID, dto.Title, dto.Content)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, article)
}

// Update — редактирование опубликованной статьи
func (h *ArticleHandler) Update(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.CtxUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	authorID := userIDVal.(uuid.UUID)

	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var dto articleUpsertDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.articleService.Update(c.Request.Context(), authorID, articleID, dto.Title, dto.Content)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, article)
}

func (h *ArticleHandler) GetByID(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	article, err := h.articleService.GetByID(c.Request.Context(), articleID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	article, err := h.articleService.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *ArticleHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	articles, err := h.articleService.List(c.Request.Context(), limit, offset)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, articles)
}
