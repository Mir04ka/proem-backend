package repository

import (
	"context"
	"errors"

	"proem-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type articleRepo struct {
	db *pgxpool.Pool
}

func NewArticleRepository(db *pgxpool.Pool) domain.ArticleRepository {
	return &articleRepo{db: db}
}

func (r *articleRepo) Create(ctx context.Context, a *domain.Article) error {
	query := `
		INSERT INTO articles (id, author_id, title, slug, content, is_published, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		a.ID, a.AuthorID, a.Title, a.Slug, a.Content, a.IsPublished, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *articleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	query := `
		SELECT a.id, a.author_id, u.email, a.title, a.slug, a.content, a.is_published, a.created_at, a.updated_at
		FROM articles a
		JOIN users u ON a.author_id = u.id
		WHERE a.id = $1
	`
	a := &domain.Article{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.AuthorID, &a.AuthorEmail, &a.Title, &a.Slug, &a.Content,
		&a.IsPublished, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *articleRepo) GetBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	query := `
		SELECT a.id, a.author_id, u.email, a.title, a.slug, a.content, a.is_published, a.created_at, a.updated_at
		FROM articles a
		JOIN users u ON a.author_id = u.id
		WHERE a.slug = $1
	`
	a := &domain.Article{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&a.ID, &a.AuthorID, &a.AuthorEmail, &a.Title, &a.Slug, &a.Content,
		&a.IsPublished, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *articleRepo) Update(ctx context.Context, a *domain.Article) error {
	query := `
		UPDATE articles
		SET title = $2, slug = $3, content = $4, is_published = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, a.ID, a.Title, a.Slug, a.Content, a.IsPublished, a.UpdatedAt)
	return err
}

func (r *articleRepo) List(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	query := `
		SELECT a.id, a.author_id, u.email, a.title, a.slug, a.content, a.is_published, a.created_at, a.updated_at
		FROM articles a
		JOIN users u ON a.author_id = u.id
		WHERE a.is_published = TRUE
		ORDER BY a.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]*domain.Article, 0)
	for rows.Next() {
		a := &domain.Article{}
		if err := rows.Scan(
			&a.ID, &a.AuthorID, &a.AuthorEmail, &a.Title, &a.Slug, &a.Content,
			&a.IsPublished, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}
