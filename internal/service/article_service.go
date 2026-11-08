package service

import (
	"context"
	"time"

	"proem-backend/internal/domain"

	"github.com/google/uuid"
)

type ArticleService interface {
	Publish(ctx context.Context, authorID uuid.UUID, title, content string, tags []string) (*domain.Article, error)
	Update(ctx context.Context, authorID, articleID uuid.UUID, title, content string, tags []string) (*domain.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Article, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Article, error)
	Vote(ctx context.Context, articleID, userID uuid.UUID, value int16) (int, error)
	AddComment(ctx context.Context, articleID, authorID uuid.UUID, content string) (*domain.Comment, error)
	GetComments(ctx context.Context, articleID uuid.UUID) ([]*domain.Comment, error)
}

type articleService struct {
	repo domain.ArticleRepository
}

func NewArticleService(repo domain.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

func (s *articleService) Publish(ctx context.Context, authorID uuid.UUID, title, content string, tags []string) (*domain.Article, error) {
	article := &domain.Article{
		ID:          uuid.New(),
		AuthorID:    authorID,
		Title:       title,
		Content:     content,
		IsPublished: true,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, article, tags); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *articleService) Update(ctx context.Context, authorID, articleID uuid.UUID, title, content string, tags []string) (*domain.Article, error) {
	article, err := s.repo.GetByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	if article.AuthorID != authorID {
		return nil, domain.ErrForbidden
	}

	article.Title = title
	article.Content = content
	article.Tags = tags
	article.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, article, tags); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *articleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) List(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *articleService) ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.ListByAuthorID(ctx, authorID, limit, offset)
}

func (s *articleService) Vote(ctx context.Context, articleID, userID uuid.UUID, value int16) (int, error) {
	return s.repo.Vote(ctx, articleID, userID, value)
}

func (s *articleService) AddComment(ctx context.Context, articleID, authorID uuid.UUID, content string) (*domain.Comment, error) {
	comment := &domain.Comment{
		ID:        uuid.New(),
		ArticleID: articleID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *articleService) GetComments(ctx context.Context, articleID uuid.UUID) ([]*domain.Comment, error) {
	return s.repo.GetComments(ctx, articleID)
}
