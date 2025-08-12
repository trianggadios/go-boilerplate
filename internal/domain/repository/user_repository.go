package repository

import (
	"boilerplate-go/internal/domain/entity"
	"context"
)

// UserRepository defines the contract for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int) error
}
