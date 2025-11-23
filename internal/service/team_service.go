package service

import (
	"context"
	"review-service/internal/domain"
	"review-service/internal/repository"
)

type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *teamService) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {

	if err := team.Validate(); err != nil {
		return nil, err
	}

	exists, err := s.teamRepo.Exists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrTeamExists
	}

	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	for _, member := range team.Members {
		user := domain.NewUser(member.UserID, member.Username, team.TeamName, member.IsActive)
		if err := user.Validate(); err != nil {
			return nil, err
		}

		if err := s.userRepo.Upsert(ctx, user); err != nil {
			return nil, err
		}
	}

	return team, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, domain.ErrTeamNotFound
	}

	return team, nil
}
