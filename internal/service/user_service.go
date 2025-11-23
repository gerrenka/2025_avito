package service

import (
	"context"
	"review-service/internal/domain"
	"review-service/internal/repository"
)

type userService struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
}

func NewUserService(userRepo repository.UserRepository, prRepo repository.PullRequestRepository) UserService {
	return &userService{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *userService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	user.SetActive(isActive)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetReviews(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {

	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		result = append(result, pr.ToShort())
	}

	return result, nil
}
