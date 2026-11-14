package service

import (
	"context"
	"strings"
	"time"

	"proem-backend/internal/domain"

	"github.com/google/uuid"
)

type ArticleService interface {
	Publish(ctx context.Context, authorID uuid.UUID, title, content string, tags []string, isPublic bool) (*domain.Article, error)
	Update(ctx context.Context, authorID, articleID uuid.UUID, title, content string, tags []string, isPublic *bool) (*domain.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error)
	List(ctx context.Context, limit, offset int, query, tag string) ([]*domain.Article, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Article, error)
	Vote(ctx context.Context, articleID, userID uuid.UUID, value int16) (int, error)
	AddComment(ctx context.Context, articleID, authorID uuid.UUID, parentID *uuid.UUID, content string) (*domain.Comment, error)
	GetComments(ctx context.Context, articleID uuid.UUID) ([]*domain.Comment, error)

	// Admin only
	ListModerationRequests(ctx context.Context, limit, offset int) ([]*domain.Article, error)
	ApprovePublic(ctx context.Context, articleID uuid.UUID) error
	RejectPublic(ctx context.Context, articleID uuid.UUID, reason *string) error
	Unpublish(ctx context.Context, articleID uuid.UUID, reason *string) error
}

type articleService struct {
	repo domain.ArticleRepository
}

func NewArticleService(repo domain.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

func (s *articleService) Publish(ctx context.Context, authorID uuid.UUID, title, content string, tags []string, isPublic bool) (*domain.Article, error) {
	modStatus := domain.ModerationNone
	if isPublic {
		// При публикации для всех статья сразу доступна по ссылке, а в общую ленту попадёт после апрува
		modStatus = domain.ModerationPending
	}

	now := time.Now()
	article := &domain.Article{
		ID:               uuid.New(),
		AuthorID:         authorID,
		Title:            title,
		Preview:          GeneratePreview(content, 250),
		Content:          content,
		IsPublished:      true,
		Visibility:       domain.VisibilityUnlisted,
		ModerationStatus: modStatus,
		Tags:             tags,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.Create(ctx, article, tags); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *articleService) Update(ctx context.Context, authorID, articleID uuid.UUID, title, content string, tags []string, isPublic *bool) (*domain.Article, error) {
	article, err := s.repo.GetByID(ctx, articleID)
	if err != nil {
		return nil, err
	}

	if article.AuthorID != authorID {
		return nil, domain.ErrForbidden
	}

	article.Title = title
	article.Content = content
	article.Preview = GeneratePreview(content, 250)
	article.Tags = tags
	article.UpdatedAt = time.Now()

	if isPublic != nil {
		if *isPublic {
			if article.Visibility != domain.VisibilityPublic {
				article.Visibility = domain.VisibilityUnlisted
				article.ModerationStatus = domain.ModerationPending
				article.ModerationReason = nil
			}
		} else {
			article.Visibility = domain.VisibilityUnlisted
			article.ModerationStatus = domain.ModerationNone
			article.ModerationReason = nil
		}
	}

	if err := s.repo.Update(ctx, article, tags); err != nil {
		return nil, err
	}
	return article, nil
}

func (s *articleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.IncrementViews(ctx, id); err == nil {
		article.ViewsCount++
	}
	if article.Preview == "" {
		article.Preview = GeneratePreview(article.Content, 250)
	}
	return article, nil
}

func (s *articleService) List(ctx context.Context, limit, offset int, query, tag string) ([]*domain.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	articles, err := s.repo.List(ctx, limit, offset, strings.TrimSpace(query), strings.TrimSpace(tag))
	if err != nil {
		return nil, err
	}
	for _, a := range articles {
		if a.Preview == "" {
			a.Preview = GeneratePreview(a.Content, 250)
		}
	}
	return articles, nil
}

func (s *articleService) ListByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	articles, err := s.repo.ListByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, a := range articles {
		if a.Preview == "" {
			a.Preview = GeneratePreview(a.Content, 250)
		}
	}
	return articles, nil
}

func (s *articleService) Vote(ctx context.Context, articleID, userID uuid.UUID, value int16) (int, error) {
	if _, err := s.repo.GetByID(ctx, articleID); err != nil {
		return 0, err
	}
	return s.repo.Vote(ctx, articleID, userID, value)
}

func (s *articleService) AddComment(ctx context.Context, articleID, authorID uuid.UUID, parentID *uuid.UUID, content string) (*domain.Comment, error) {
	if _, err := s.repo.GetByID(ctx, articleID); err != nil {
		return nil, err
	}

	if parentID != nil {
		parent, err := s.repo.GetCommentByID(ctx, *parentID)
		if err != nil {
			return nil, domain.ErrInvalidParentComment
		}
		if parent.ArticleID != articleID {
			return nil, domain.ErrInvalidParentComment
		}
	}

	now := time.Now()
	comment := &domain.Comment{
		ID:        uuid.New(),
		ArticleID: articleID,
		ParentID:  parentID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *articleService) GetComments(ctx context.Context, articleID uuid.UUID) ([]*domain.Comment, error) {
	if _, err := s.repo.GetByID(ctx, articleID); err != nil {
		return nil, err
	}

	flat, err := s.repo.GetComments(ctx, articleID)
	if err != nil {
		return nil, err
	}

	// Собираем дерево комментариев за O(N)
	byID := make(map[uuid.UUID]*domain.Comment, len(flat))
	for _, c := range flat {
		c.Replies = make([]*domain.Comment, 0)
		byID[c.ID] = c
	}

	roots := make([]*domain.Comment, 0)
	for _, c := range flat {
		if c.ParentID != nil {
			if parent, ok := byID[*c.ParentID]; ok {
				parent.Replies = append(parent.Replies, c)
				continue
			}
		}
		roots = append(roots, c)
	}
	return roots, nil
}

// --- Админ-методы ---

func (s *articleService) ListModerationRequests(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPendingModeration(ctx, limit, offset)
}

func (s *articleService) ApprovePublic(ctx context.Context, articleID uuid.UUID) error {
	article, err := s.repo.GetByIDAnyStatus(ctx, articleID)
	if err != nil {
		return err
	}
	if !article.IsPublished || article.ModerationStatus != domain.ModerationPending {
		return domain.ErrArticleNotPending
	}
	return s.repo.UpdateModeration(ctx, articleID, true, domain.VisibilityPublic, domain.ModerationApproved, nil)
}

func (s *articleService) RejectPublic(ctx context.Context, articleID uuid.UUID, reason *string) error {
	article, err := s.repo.GetByIDAnyStatus(ctx, articleID)
	if err != nil {
		return err
	}
	if !article.IsPublished || article.ModerationStatus != domain.ModerationPending {
		return domain.ErrArticleNotPending
	}
	// Статья остаётся доступной по ссылке, но не попадает в общий список
	return s.repo.UpdateModeration(ctx, articleID, true, domain.VisibilityUnlisted, domain.ModerationRejected, reason)
}

func (s *articleService) Unpublish(ctx context.Context, articleID uuid.UUID, reason *string) error {
	if _, err := s.repo.GetByIDAnyStatus(ctx, articleID); err != nil {
		return err
	}
	// Полное снятие публикации: доступ закрывается даже по прямой ссылке
	return s.repo.UpdateModeration(ctx, articleID, false, domain.VisibilityUnlisted, domain.ModerationRejected, reason)
}
