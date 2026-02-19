package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yumikokawaii/sherry-archive/internal/model"
	"github.com/yumikokawaii/sherry-archive/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type UpdateUserInput struct {
	Bio      *string
	Username *string
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) Update(ctx context.Context, userID uuid.UUID, in UpdateUserInput) (*model.User, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if in.Bio != nil {
		u.Bio = *in.Bio
	}
	if in.Username != nil {
		u.Username = *in.Username
	}
	u.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) (*model.User, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.AvatarURL = avatarURL
	u.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
