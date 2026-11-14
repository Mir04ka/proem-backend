package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"time"

	"proem-backend/internal/domain"
	"proem-backend/pkg/token"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) error
	VerifyEmail(ctx context.Context, verificationToken string) error
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error)
	GetProfileByUsername(ctx context.Context, username string) (*domain.UserProfile, error)
	ListUsers(ctx context.Context, limit, offset int) ([]*domain.UserProfile, error)
}

type authService struct {
	userRepo domain.UserRepository
	jwtMgr   *token.JWTManager
	mailer   MailerService
}

func NewAuthService(repo domain.UserRepository, jwtMgr *token.JWTManager, mailer MailerService) AuthService {
	return &authService{userRepo: repo, jwtMgr: jwtMgr, mailer: mailer}
}

func (s *authService) Register(ctx context.Context, username, email, password string) error {
	if !usernameRegex.MatchString(username) {
		return domain.ErrInvalidUsername
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	verToken := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(24 * time.Hour)

	user := &domain.User{
		ID:                    uuid.New(),
		Username:              username,
		Email:                 email,
		PasswordHash:          string(hash),
		Role:                  domain.RoleUser,
		IsVerified:            false,
		VerificationToken:     &verToken,
		VerificationExpiresAt: &expiresAt,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	go func() {
		_ = s.mailer.SendVerificationEmail(email, verToken)
	}()

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, verToken string) error {
	user, err := s.userRepo.GetByVerificationToken(ctx, verToken)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if user.VerificationExpiresAt == nil || time.Now().After(*user.VerificationExpiresAt) {
		return domain.ErrInvalidToken
	}

	user.IsVerified = true
	user.VerificationToken = nil
	user.VerificationExpiresAt = nil
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

func (s *authService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	if !user.IsVerified {
		return "", nil, domain.ErrEmailNotVerified
	}

	tokenStr, err := s.jwtMgr.Generate(user.ID, user.Email)
	if err != nil {
		return "", nil, err
	}

	return tokenStr, user, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return domain.ErrInvalidOldPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(newHash)
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.UserProfile{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) GetProfileByUsername(ctx context.Context, username string) (*domain.UserProfile, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return &domain.UserProfile{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) ListUsers(ctx context.Context, limit, offset int) ([]*domain.UserProfile, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.userRepo.List(ctx, limit, offset)
}
