package repository

import (
	"context"
	"errors"

	"proem-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, is_verified, verification_token, verification_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		u.ID, u.Username, u.Email, u.PasswordHash, u.IsVerified,
		u.VerificationToken, u.VerificationExpiresAt, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				return domain.ErrDuplicateUsername
			}
			return domain.ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_verified, verification_token, verification_expires_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsVerified,
		&u.VerificationToken, &u.VerificationExpiresAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_verified, verification_token, verification_expires_at, created_at, updated_at
		FROM users WHERE email = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsVerified,
		&u.VerificationToken, &u.VerificationExpiresAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_verified, verification_token, verification_expires_at, created_at, updated_at
		FROM users WHERE username = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsVerified,
		&u.VerificationToken, &u.VerificationExpiresAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepo) GetByVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_verified, verification_token, verification_expires_at, created_at, updated_at
		FROM users WHERE verification_token = $1
	`
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, token).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.IsVerified,
		&u.VerificationToken, &u.VerificationExpiresAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, password_hash = $4, is_verified = $5,
		    verification_token = $6, verification_expires_at = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		u.ID, u.Username, u.Email, u.PasswordHash, u.IsVerified,
		u.VerificationToken, u.VerificationExpiresAt, u.UpdatedAt,
	)
	return err
}
