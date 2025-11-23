package postgres

import (
	"context"
	"database/sql"

	"review-service/internal/domain"
	"time"

	"github.com/lib/pq"
)

type TeamPostgresRepository struct {
	db *sql.DB
}

func NewTeamPostgresRepository(db *sql.DB) *TeamPostgresRepository {
	return &TeamPostgresRepository{
		db: db,
	}
}

func (r *TeamPostgresRepository) Create(ctx context.Context, team *domain.Team) error {
	query := `
		INSERT INTO teams (team_name, created_at, updated_at)
		VALUES ($1, $2, $3)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, team.TeamName, now, now)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrTeamExists
		}
		return err
	}

	team.CreatedAt = now
	team.UpdatedAt = now

	return nil
}

func (r *TeamPostgresRepository) GetByName(ctx context.Context, teamName string) (*domain.Team, error) {
	var team domain.Team
	var createdAt, updatedAt time.Time

	teamQuery := `
		SELECT team_name, created_at, updated_at
		FROM teams
		WHERE team_name = $1
	`

	err := r.db.QueryRowContext(ctx, teamQuery, teamName).Scan(
		&team.TeamName,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	team.CreatedAt = createdAt
	team.UpdatedAt = updatedAt

	membersQuery := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY username
	`

	rows, err := r.db.QueryContext(ctx, membersQuery, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	team.Members = []domain.TeamMember{}
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &team, nil
}

func (r *TeamPostgresRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *TeamPostgresRepository) Update(ctx context.Context, team *domain.Team) error {
	query := `
		UPDATE teams
		SET updated_at = $1
		WHERE team_name = $2
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, team.TeamName)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrTeamNotFound
	}

	team.UpdatedAt = now
	return nil
}
