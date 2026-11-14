package domain

import "errors"

var (
	ErrNotFound             = errors.New("record not found")
	ErrDuplicateEmail       = errors.New("email already in use")
	ErrDuplicateUsername    = errors.New("username already in use")
	ErrInvalidUsername      = errors.New("username must be 3-20 characters long and contain only lowercase latin letters, digits, and underscores")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrEmailNotVerified     = errors.New("email not verified, check your inbox")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrForbidden            = errors.New("you do not have permission to perform this action")
	ErrInvalidOldPassword   = errors.New("invalid current password")
	ErrInvalidFileFormat    = errors.New("invalid file format, only images allowed")
	ErrFileSizeTooLarge     = errors.New("file size exceeds maximum allowed limit")
	ErrInvalidParentComment = errors.New("invalid parent comment")
	ErrArticleNotPending    = errors.New("article is not pending moderation")
)
