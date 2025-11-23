package service

import (
	"context"
	"math/rand"
	"review-service/internal/domain"
	"review-service/internal/repository"
	"time"
)

type pullRequestService struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
	rand     *rand.Rand
}

func NewPullRequestService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
) PullRequestService {
	return &pullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *pullRequestService) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {

	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		return nil, err
	}
	if author == nil {
		return nil, domain.ErrUserNotFound
	}

	team, err := s.teamRepo.GetByName(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, domain.ErrTeamNotFound
	}

	pr := domain.NewPullRequest(prID, prName, authorID)

	reviewers := s.selectReviewers(team, authorID, 2)
	pr.AssignReviewers(reviewers)

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *pullRequestService) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, domain.ErrPRNotFound
	}

	pr.Merge()

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *pullRequestService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error) {

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}
	if pr == nil {
		return nil, "", domain.ErrPRNotFound
	}

	if err := pr.CanReassign(); err != nil {
		return nil, "", err
	}

	if !pr.HasReviewer(oldUserID) {
		return nil, "", domain.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		return nil, "", err
	}
	if oldReviewer == nil {
		return nil, "", domain.ErrUserNotFound
	}

	team, err := s.teamRepo.GetByName(ctx, oldReviewer.TeamName)
	if err != nil {
		return nil, "", err
	}
	if team == nil {
		return nil, "", domain.ErrTeamNotFound
	}

	excludeUserIDs := append([]string{pr.AuthorID}, pr.AssignedReviewers...)

	candidates := team.GetActiveMembersExcluding(excludeUserIDs...)

	if len(candidates) == 0 {
		return nil, "", domain.ErrNoCandidate
	}

	newReviewer := candidates[s.rand.Intn(len(candidates))]

	if err := pr.ReplaceReviewer(oldUserID, newReviewer.UserID); err != nil {
		return nil, "", err
	}

	if err := s.prRepo.Update(ctx, pr); err != nil {
		return nil, "", err
	}

	return pr, newReviewer.UserID, nil
}

func (s *pullRequestService) selectReviewers(team *domain.Team, authorID string, maxCount int) []string {

	candidates := team.GetActiveMembersExcluding(authorID)

	if len(candidates) == 0 || maxCount == 0 {
		return []string{}
	}

	count := maxCount
	if len(candidates) < count {
		count = len(candidates)
	}

	shuffled := make([]domain.TeamMember, len(candidates))
	copy(shuffled, candidates)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := s.rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = shuffled[i].UserID
	}

	return reviewers
}
