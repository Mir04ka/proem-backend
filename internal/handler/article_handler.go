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
	Title   string   `json:"title" binding:"required,min=3"`
	Content string   `json:"content" binding:"required"`
	Tags    []string `json:"tags"`
}

func (h *ArticleHandler) Publish(c *gin.Context) {
	authorID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)

	var dto articleUpsertDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.articleService.Publish(c.Request.Context(), authorID, dto.Title, dto.Content, dto.Tags)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, article)
}

func (h *ArticleHandler) Update(c *gin.Context) {
	authorID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)
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

	article, err := h.articleService.Update(c.Request.Context(), authorID, articleID, dto.Title, dto.Content, dto.Tags)
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

func (h *ArticleHandler) ListByUser(c *gin.Context) {
	authorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	articles, err := h.articleService.ListByAuthor(c.Request.Context(), authorID, limit, offset)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, articles)
}

// POST /api/v1/articles/:id/vote
type voteDTO struct {
	Value int16 `json:"value" binding:"required,oneof=1 -1"`
}

func (h *ArticleHandler) Vote(c *gin.Context) {
	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var dto voteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Value must be 1 or -1"})
		return
	}

	score, err := h.articleService.Vote(c.Request.Context(), articleID, userID, dto.Value)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"rating_score": score})
}

// GET /api/v1/articles/:id/comments
func (h *ArticleHandler) GetComments(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	comments, err := h.articleService.GetComments(c.Request.Context(), articleID)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, comments)
}

// POST /api/v1/articles/:id/comments
type commentDTO struct {
	Content string `json:"content" binding:"required,min=1"`
}

func (h *ArticleHandler) AddComment(c *gin.Context) {
	userID := c.MustGet(middleware.CtxUserIDKey).(uuid.UUID)
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var dto commentDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.articleService.AddComment(c.Request.Context(), articleID, userID, dto.Content)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}
