package domain

import "time"

type Team struct {
	TeamName  string       `json:"team_name" db:"team_name"`
	Members   []TeamMember `json:"members"`
	CreatedAt time.Time    `json:"-" db:"created_at"`
	UpdatedAt time.Time    `json:"-" db:"updated_at"`
}

func NewTeam(teamName string, members []TeamMember) *Team {
	now := time.Now()
	return &Team{
		TeamName:  teamName,
		Members:   members,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Team) Validate() error {
	if t.TeamName == "" {
		return NewDomainError(ErrCodeNotFound, "team_name cannot be empty")
	}
	if len(t.Members) == 0 {
		return NewDomainError(ErrCodeNotFound, "team must have at least one member")
	}

	userIDs := make(map[string]bool)
	for _, member := range t.Members {
		if member.UserID == "" {
			return NewDomainError(ErrCodeNotFound, "member user_id cannot be empty")
		}
		if member.Username == "" {
			return NewDomainError(ErrCodeNotFound, "member username cannot be empty")
		}

		if userIDs[member.UserID] {
			return NewDomainError(ErrCodeNotFound, "duplicate user_id in team members")
		}
		userIDs[member.UserID] = true
	}

	return nil
}

func (t *Team) GetActiveMembers() []TeamMember {
	activeMembers := make([]TeamMember, 0)
	for _, member := range t.Members {
		if member.IsActive {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}

func (t *Team) GetActiveMembersExcluding(excludeUserIDs ...string) []TeamMember {
	excludeMap := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}

	activeMembers := make([]TeamMember, 0)
	for _, member := range t.Members {
		if member.IsActive && !excludeMap[member.UserID] {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}

func (t *Team) HasMember(userID string) bool {
	for _, member := range t.Members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}
