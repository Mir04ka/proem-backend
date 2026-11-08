package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID          uuid.UUID `json:"id"`
	AuthorID    uuid.UUID `json:"author_id"`
	AuthorEmail string    `json:"author_email,omitempty"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Content     string    `json:"content"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ArticleRepository interface {
	Create(ctx context.Context, article *Article) error
	GetByID(ctx context.Context, id uuid.UUID) (*Article, error)
	GetBySlug(ctx context.Context, slug string) (*Article, error)
	Update(ctx context.Context, article *Article) error
	List(ctx context.Context, limit, offset int) ([]*Article, error)
}
