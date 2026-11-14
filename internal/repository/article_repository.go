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

const articleSelectFields = `
	SELECT a.id, a.author_id, u.username, u.email, a.title, a.preview, a.content,
	       a.is_published, a.visibility, a.moderation_status, a.moderation_reason,
	       (SELECT COALESCE(SUM(value), 0) FROM article_votes WHERE article_id = a.id) AS rating,
	       COALESCE(array_remove(array_agg(DISTINCT t.name), NULL), '{}') AS tags,
	       a.created_at, a.updated_at
	FROM articles a
	JOIN users u ON a.author_id = u.id
	LEFT JOIN article_tags at ON a.id = at.article_id
	LEFT JOIN tags t ON at.tag_id = t.id
`

func scanArticle(row pgx.Row) (*domain.Article, error) {
	a := &domain.Article{}
	err := row.Scan(
		&a.ID, &a.AuthorID, &a.AuthorUsername, &a.AuthorEmail, &a.Title, &a.Preview, &a.Content,
		&a.IsPublished, &a.Visibility, &a.ModerationStatus, &a.ModerationReason,
		&a.Rating, &a.Tags, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (r *articleRepo) Create(ctx context.Context, a *domain.Article, tags []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO articles (id, author_id, title, preview, content, is_published, visibility, moderation_status, moderation_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(ctx, query,
		a.ID, a.AuthorID, a.Title, a.Preview, a.Content,
		a.IsPublished, a.Visibility, a.ModerationStatus, a.ModerationReason,
		a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if err := r.syncTags(ctx, tx, a.ID, tags); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *articleRepo) Update(ctx context.Context, a *domain.Article, tags []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE articles
		SET title = $2, preview = $3, content = $4, visibility = $5,
		    moderation_status = $6, moderation_reason = $7, updated_at = $8
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, query,
		a.ID, a.Title, a.Preview, a.Content,
		a.Visibility, a.ModerationStatus, a.ModerationReason, a.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM article_tags WHERE article_id = $1`, a.ID); err != nil {
		return err
	}
	if err := r.syncTags(ctx, tx, a.ID, tags); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *articleRepo) UpdateModeration(ctx context.Context, id uuid.UUID, isPublished bool, visibility, moderationStatus string, reason *string) error {
	query := `
		UPDATE articles
		SET is_published = $2, visibility = $3, moderation_status = $4, moderation_reason = $5, updated_at = NOW()
		WHERE id = $1
	`
	cmd, err := r.db.Exec(ctx, query, id, isPublished, visibility, moderationStatus, reason)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *articleRepo) syncTags(ctx context.Context, tx pgx.Tx, articleID uuid.UUID, tags []string) error {
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		var tagID uuid.UUID
		tagQuery := `
			INSERT INTO tags (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`
		if err := tx.QueryRow(ctx, tagQuery, tag).Scan(&tagID); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO article_tags (article_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, articleID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetByID возвращает только опубликованные статьи (публичные или доступные по ссылке).
// Снятые админом (is_published = FALSE) недоступны даже по прямой ссылке.
func (r *articleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	query := articleSelectFields + `
		WHERE a.id = $1 AND a.is_published = TRUE
		GROUP BY a.id, u.username, u.email
	`
	a, err := scanArticle(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

// GetByIDAnyStatus используется админом для модерации.
func (r *articleRepo) GetByIDAnyStatus(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	query := articleSelectFields + `
		WHERE a.id = $1
		GROUP BY a.id, u.username, u.email
	`
	a, err := scanArticle(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func (r *articleRepo) List(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	query := articleSelectFields + `
		WHERE a.is_published = TRUE AND a.visibility = 'public' AND a.moderation_status = 'approved'
		GROUP BY a.id, u.username, u.email
		ORDER BY a.created_at DESC
		LIMIT $1 OFFSET $2
	`
	return r.fetchArticles(ctx, query, limit, offset)
}

func (r *articleRepo) ListByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Article, error) {
	query := articleSelectFields + `
		WHERE a.author_id = $1 AND a.is_published = TRUE AND a.visibility = 'public' AND a.moderation_status = 'approved'
		GROUP BY a.id, u.username, u.email
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.fetchArticles(ctx, query, authorID, limit, offset)
}

func (r *articleRepo) ListPendingModeration(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	query := articleSelectFields + `
		WHERE a.is_published = TRUE AND a.moderation_status = 'pending'
		GROUP BY a.id, u.username, u.email
		ORDER BY a.created_at ASC
		LIMIT $1 OFFSET $2
	`
	return r.fetchArticles(ctx, query, limit, offset)
}

func (r *articleRepo) fetchArticles(ctx context.Context, query string, args ...interface{}) ([]*domain.Article, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]*domain.Article, 0)
	for rows.Next() {
		a, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (r *articleRepo) Vote(ctx context.Context, articleID, userID uuid.UUID, value int16) (int, error) {
	voteQuery := `
		INSERT INTO article_votes (article_id, user_id, value, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (article_id, user_id) DO UPDATE SET value = EXCLUDED.value
	`
	if _, err := r.db.Exec(ctx, voteQuery, articleID, userID, value); err != nil {
		return 0, err
	}

	var score int
	sumQuery := `SELECT COALESCE(SUM(value), 0) FROM article_votes WHERE article_id = $1`
	err := r.db.QueryRow(ctx, sumQuery, articleID).Scan(&score)
	return score, err
}

func (r *articleRepo) AddComment(ctx context.Context, c *domain.Comment) error {
	query := `
		INSERT INTO comments (id, article_id, parent_id, author_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, c.ID, c.ArticleID, c.ParentID, c.AuthorID, c.Content, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return err
	}

	return r.db.QueryRow(ctx, `SELECT username FROM users WHERE id = $1`, c.AuthorID).Scan(&c.AuthorUsername)
}

func (r *articleRepo) GetCommentByID(ctx context.Context, commentID uuid.UUID) (*domain.Comment, error) {
	query := `
		SELECT c.id, c.article_id, c.parent_id, c.author_id, u.username, c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.id = $1
	`
	c := &domain.Comment{}
	err := r.db.QueryRow(ctx, query, commentID).Scan(
		&c.ID, &c.ArticleID, &c.ParentID, &c.AuthorID, &c.AuthorUsername, &c.Content, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return c, nil
}

func (r *articleRepo) GetComments(ctx context.Context, articleID uuid.UUID) ([]*domain.Comment, error) {
	query := `
		SELECT c.id, c.article_id, c.parent_id, c.author_id, u.username, c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.article_id = $1
		ORDER BY c.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*domain.Comment, 0)
	for rows.Next() {
		c := &domain.Comment{}
		if err := rows.Scan(
			&c.ID, &c.ArticleID, &c.ParentID, &c.AuthorID, &c.AuthorUsername, &c.Content, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}
