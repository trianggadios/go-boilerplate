package errors

import "errors"

// Common application errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInternalServer     = errors.New("internal server error")
)

// IsUserNotFound checks if the error is a user not found error.
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
