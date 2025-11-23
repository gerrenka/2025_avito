package service

import (
	"context"
	"review-service/internal/domain"
)

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetReviews(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

type TeamService interface {
	CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error)
}
