package repository

import (
	"context"
	"review-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByTeam(ctx context.Context, teamName string) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Upsert(ctx context.Context, user *domain.User) error
	Exists(ctx context.Context, userID string) (bool, error)
}

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, teamName string) (*domain.Team, error)
	Exists(ctx context.Context, teamName string) (bool, error)
	Update(ctx context.Context, team *domain.Team) error
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	GetByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error)
	Update(ctx context.Context, pr *domain.PullRequest) error
	Exists(ctx context.Context, prID string) (bool, error)
}
