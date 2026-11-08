package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	PasswordHash          string     `json:"-"`
	IsVerified            bool       `json:"is_verified"`
	VerificationToken     *string    `json:"-"`
	VerificationExpiresAt *time.Time `json:"-"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByVerificationToken(ctx context.Context, token string) (*User, error)
	Update(ctx context.Context, user *User) error
}
