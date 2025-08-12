package user

import (
	"boilerplate-go/internal/domain/entity"
	"boilerplate-go/internal/domain/repository"
	"context"
)

type UserUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (uc *UserUsecase) GetProfile(ctx context.Context, userID int) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

func (uc *UserUsecase) UpdateProfile(ctx context.Context, user *entity.User) error {
	return uc.userRepo.Update(ctx, user)
}
