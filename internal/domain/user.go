package domain

import "time"

type User struct {
	UserID    string    `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	TeamName  string    `json:"team_name" db:"team_name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func NewUser(userID, username, teamName string, isActive bool) *User {
	now := time.Now()
	return &User{
		UserID:    userID,
		Username:  username,
		TeamName:  teamName,
		IsActive:  isActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) SetActive(isActive bool) {
	u.IsActive = isActive
	u.UpdatedAt = time.Now()
}

func (u *User) ToTeamMember() TeamMember {
	return TeamMember{
		UserID:   u.UserID,
		Username: u.Username,
		IsActive: u.IsActive,
	}
}

func (u *User) Validate() error {
	if u.UserID == "" {
		return NewDomainError(ErrCodeNotFound, "user_id cannot be empty")
	}
	if u.Username == "" {
		return NewDomainError(ErrCodeNotFound, "username cannot be empty")
	}
	if u.TeamName == "" {
		return NewDomainError(ErrCodeNotFound, "team_name cannot be empty")
	}
	return nil
}
