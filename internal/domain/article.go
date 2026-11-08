package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID             uuid.UUID `json:"id"`
	AuthorID       uuid.UUID `json:"author_id"`
	AuthorUsername string    `json:"author_username,omitempty"`
	AuthorEmail    string    `json:"author_email,omitempty"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	IsPublished    bool      `json:"is_published"`
	Rating         int       `json:"rating"`
	Tags           []string  `json:"tags"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Comment struct {
	ID             uuid.UUID `json:"id"`
	ArticleID      uuid.UUID `json:"article_id"`
	AuthorID       uuid.UUID `json:"author_id"`
	AuthorUsername string    `json:"author_username"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ArticleRepository interface {
	Create(ctx context.Context, article *Article, tags []string) error
	GetByID(ctx context.Context, id uuid.UUID) (*Article, error)
	Update(ctx context.Context, article *Article, tags []string) error
	List(ctx context.Context, limit, offset int) ([]*Article, error)
	ListByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*Article, error)

	Vote(ctx context.Context, articleID, userID uuid.UUID, value int16) (int, error)

	AddComment(ctx context.Context, comment *Comment) error
	GetComments(ctx context.Context, articleID uuid.UUID) ([]*Comment, error)
}
