package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	       a.views_count,
	       (SELECT COUNT(*) FROM comments WHERE article_id = a.id) AS comments_count,
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
		&a.ViewsCount, &a.CommentsCount, &a.Rating, &a.Tags, &a.CreatedAt, &a.UpdatedAt,
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

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *articleRepo) List(ctx context.Context, limit, offset int, searchQuery, tag string) ([]*domain.Article, error) {
	conds := []string{
		"a.is_published = TRUE",
		"a.visibility = 'public'",
		"a.moderation_status = 'approved'",
	}
	args := make([]interface{}, 0, 4)

	if searchQuery != "" {
		args = append(args, "%"+escapeLike(searchQuery)+"%")
		idx := len(args)
		conds = append(conds, fmt.Sprintf(`(
			a.title ILIKE $%d ESCAPE '\'
			OR EXISTS (
				SELECT 1 FROM article_tags at2
				JOIN tags t2 ON at2.tag_id = t2.id
				WHERE at2.article_id = a.id AND t2.name ILIKE $%d ESCAPE '\'
			)
		)`, idx, idx))
	}

	if tag != "" {
		args = append(args, tag)
		idx := len(args)
		conds = append(conds, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM article_tags at3
			JOIN tags t3 ON at3.tag_id = t3.id
			WHERE at3.article_id = a.id AND LOWER(t3.name) = LOWER($%d)
		)`, idx))
	}

	args = append(args, limit, offset)
	limitIdx := len(args) - 1
	offsetIdx := len(args)

	query := articleSelectFields +
		" WHERE " + strings.Join(conds, " AND ") +
		fmt.Sprintf(" GROUP BY a.id, u.username, u.email ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", limitIdx, offsetIdx)

	return r.fetchArticles(ctx, query, args...)
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

func (r *articleRepo) IncrementViews(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE articles
		SET views_count = views_count + 1
		WHERE id = $1 AND is_published = TRUE
	`, id)
	return err
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
