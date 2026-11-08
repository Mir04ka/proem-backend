package domain

import "errors"

var (
	ErrNotFound           = errors.New("record not found")
	ErrDuplicateEmail     = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailNotVerified   = errors.New("email not verified, check your inbox")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrForbidden          = errors.New("you do not have permission to perform this action")
)
