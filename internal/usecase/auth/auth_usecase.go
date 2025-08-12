package auth

import (
	"boilerplate-go/config"
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/internal/domain/repository"
	"boilerplate-go/pkg/errors"
	"boilerplate-go/pkg/hash"
	"boilerplate-go/pkg/jwt"
	"context"
	"fmt"
)

// AuthUsecase handles authentication business logic.
type AuthUsecase struct {
	userRepo  repository.UserRepository
	jwtConfig config.JWTConfig
}

// NewAuthUsecase creates a new authentication use case.
func NewAuthUsecase(userRepo repository.UserRepository, jwtConfig config.JWTConfig) *AuthUsecase {
	return &AuthUsecase{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, req *entity.RegisterRequest) (*entity.User, error) {
	existingUser, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.IsUserNotFound(err) {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return nil, errors.ErrUserAlreadyExists
	}

	existingUser, err = uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.IsUserNotFound(err) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, errors.ErrUserAlreadyExists
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, req *entity.LoginRequest) (*entity.LoginResponse, error) {
	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.IsUserNotFound(err) {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !hash.CheckPassword(req.Password, user.Password) {
		return nil, errors.ErrInvalidCredentials
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, uc.jwtConfig.SecretKey, uc.jwtConfig.ExpiryTime)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &entity.LoginResponse{
		Token: token,
		User:  user,
	}, nil
}
