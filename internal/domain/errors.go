package domain

import "fmt"

type ErrorCode string

const (
	ErrCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrCodePRExists    ErrorCode = "PR_EXISTS"
	ErrCodePRMerged    ErrorCode = "PR_MERGED"
	ErrCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrTeamExists = &DomainError{
		Code:    ErrCodeTeamExists,
		Message: "team_name already exists",
	}

	ErrPRExists = &DomainError{
		Code:    ErrCodePRExists,
		Message: "PR id already exists",
	}

	ErrPRMerged = &DomainError{
		Code:    ErrCodePRMerged,
		Message: "cannot reassign on merged PR",
	}

	ErrNotAssigned = &DomainError{
		Code:    ErrCodeNotAssigned,
		Message: "reviewer is not assigned to this PR",
	}

	ErrNoCandidate = &DomainError{
		Code:    ErrCodeNoCandidate,
		Message: "no active replacement candidate in team",
	}

	ErrTeamNotFound = &DomainError{
		Code:    ErrCodeNotFound,
		Message: "team not found",
	}

	ErrUserNotFound = &DomainError{
		Code:    ErrCodeNotFound,
		Message: "user not found",
	}

	ErrPRNotFound = &DomainError{
		Code:    ErrCodeNotFound,
		Message: "pull request not found",
	}
)

func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

func GetErrorCode(err error) (ErrorCode, bool) {
	if domainErr, ok := err.(*DomainError); ok {
		return domainErr.Code, true
	}
	return "", false
}
