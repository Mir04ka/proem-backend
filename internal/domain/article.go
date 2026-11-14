package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	VisibilityUnlisted = "unlisted"
	VisibilityPublic   = "public"

	ModerationNone     = "none"
	ModerationPending  = "pending"
	ModerationApproved = "approved"
	ModerationRejected = "rejected"
)

type Article struct {
	ID               uuid.UUID `json:"id"`
	AuthorID         uuid.UUID `json:"author_id"`
	AuthorUsername   string    `json:"author_username,omitempty"`
	AuthorEmail      string    `json:"author_email,omitempty"`
	Title            string    `json:"title"`
	Preview          string    `json:"preview"`
	Content          string    `json:"content"`
	IsPublished      bool      `json:"is_published"`
	Visibility       string    `json:"visibility"`
	ModerationStatus string    `json:"moderation_status"`
	ModerationReason *string   `json:"moderation_reason,omitempty"`
	ViewsCount       int64     `json:"views_count"`
	CommentsCount    int       `json:"comments_count"`
	Rating           int       `json:"rating"`
	Tags             []string  `json:"tags"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Comment struct {
	ID             uuid.UUID  `json:"id"`
	ArticleID      uuid.UUID  `json:"article_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	AuthorID       uuid.UUID  `json:"author_id"`
	AuthorUsername string     `json:"author_username"`
	Content        string     `json:"content"`
	Replies        []*Comment `json:"replies,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ArticleRepository interface {
	Create(ctx context.Context, article *Article, tags []string) error
	GetByID(ctx context.Context, id uuid.UUID) (*Article, error)
	GetByIDAnyStatus(ctx context.Context, id uuid.UUID) (*Article, error)
	Update(ctx context.Context, article *Article, tags []string) error
	UpdateModeration(ctx context.Context, id uuid.UUID, isPublished bool, visibility, moderationStatus string, reason *string) error

	List(ctx context.Context, limit, offset int, query, tag string) ([]*Article, error)
	ListByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*Article, error)
	ListPendingModeration(ctx context.Context, limit, offset int) ([]*Article, error)

	Vote(ctx context.Context, articleID, userID uuid.UUID, value int16) (int, error)
	IncrementViews(ctx context.Context, articleID uuid.UUID) error

	AddComment(ctx context.Context, comment *Comment) error
	GetCommentByID(ctx context.Context, commentID uuid.UUID) (*Comment, error)
	GetComments(ctx context.Context, articleID uuid.UUID) ([]*Comment, error)
}
