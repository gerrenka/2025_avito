package domain

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID          string     `json:"author_id" db:"author_id"`
	Status            PRStatus   `json:"status" db:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty" db:"created_at"`
	MergedAt          *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          PRStatus `json:"status"`
}

func NewPullRequest(prID, prName, authorID string) *PullRequest {
	now := time.Now()
	return &PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            PRStatusOpen,
		AssignedReviewers: []string{},
		CreatedAt:         &now,
		MergedAt:          nil,
	}
}

func (pr *PullRequest) Validate() error {
	if pr.PullRequestID == "" {
		return NewDomainError(ErrCodeNotFound, "pull_request_id cannot be empty")
	}
	if pr.PullRequestName == "" {
		return NewDomainError(ErrCodeNotFound, "pull_request_name cannot be empty")
	}
	if pr.AuthorID == "" {
		return NewDomainError(ErrCodeNotFound, "author_id cannot be empty")
	}
	if pr.Status != PRStatusOpen && pr.Status != PRStatusMerged {
		return NewDomainError(ErrCodeNotFound, "invalid PR status")
	}
	return nil
}

func (pr *PullRequest) Merge() {
	if pr.Status != PRStatusMerged {
		now := time.Now()
		pr.Status = PRStatusMerged
		pr.MergedAt = &now
	}
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == PRStatusMerged
}

func (pr *PullRequest) CanReassign() error {
	if pr.IsMerged() {
		return ErrPRMerged
	}
	return nil
}

func (pr *PullRequest) AssignReviewers(reviewerIDs []string) {

	if len(reviewerIDs) > 2 {
		reviewerIDs = reviewerIDs[:2]
	}
	pr.AssignedReviewers = reviewerIDs
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == userID {
			return true
		}
	}
	return false
}

func (pr *PullRequest) ReplaceReviewer(oldUserID, newUserID string) error {
	if err := pr.CanReassign(); err != nil {
		return err
	}

	if !pr.HasReviewer(oldUserID) {
		return ErrNotAssigned
	}

	newReviewers := make([]string, 0, len(pr.AssignedReviewers))
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldUserID {
			newReviewers = append(newReviewers, newUserID)
		} else {
			newReviewers = append(newReviewers, reviewerID)
		}
	}

	pr.AssignedReviewers = newReviewers
	return nil
}

func (pr *PullRequest) ToShort() PullRequestShort {
	return PullRequestShort{
		PullRequestID:   pr.PullRequestID,
		PullRequestName: pr.PullRequestName,
		AuthorID:        pr.AuthorID,
		Status:          pr.Status,
	}
}

func (pr *PullRequest) GetReviewerCount() int {
	return len(pr.AssignedReviewers)
}
