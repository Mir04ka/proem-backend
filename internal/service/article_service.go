package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"proem-backend/internal/domain"

	"github.com/google/uuid"
)

type ArticleService interface {
	Publish(ctx context.Context, authorID uuid.UUID, title, content string) (*domain.Article, error)
	Update(ctx context.Context, authorID, articleID uuid.UUID, title, content string) (*domain.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Article, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Article, error)
}

type articleService struct {
	repo domain.ArticleRepository
}

func NewArticleService(repo domain.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

func (s *articleService) Publish(ctx context.Context, authorID uuid.UUID, title, content string) (*domain.Article, error) {
	slug := generateSlug(title)

	article := &domain.Article{
		ID:          uuid.New(),
		AuthorID:    authorID,
		Title:       title,
		Slug:        slug,
		Content:     content,
		IsPublished: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, article); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *articleService) Update(ctx context.Context, authorID, articleID uuid.UUID, title, content string) (*domain.Article, error) {
	article, err := s.repo.GetByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	// Редактировать может только автор
	if article.AuthorID != authorID {
		return nil, domain.ErrForbidden
	}

	article.Title = title
	article.Content = content
	article.Slug = generateSlug(title)
	article.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, article); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *articleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) GetBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	return s.repo.GetBySlug(ctx, slug)
}

func (s *articleService) List(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset)
}

var nonWordRegex = regexp.MustCompile(`[^\w\s\p{Cyrillic}-]`)
var spacesRegex = regexp.MustCompile(`[\s_-]+`)

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = nonWordRegex.ReplaceAllString(slug, "")
	slug = spacesRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	// Добавляем короткий суффикс уникальности
	shortID := strings.Split(uuid.New().String(), "-")[0]
	return fmt.Sprintf("%s-%s", slug, shortID)
}
