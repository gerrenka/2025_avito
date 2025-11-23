package postgres

import (
	"context"
	"database/sql"
	"review-service/internal/domain"

	"github.com/lib/pq"
)

type PullRequestPostgresRepository struct {
	db *sql.DB
}

func NewPullRequestPostgresRepository(db *sql.DB) *PullRequestPostgresRepository {
	return &PullRequestPostgresRepository{
		db: db,
	}
}

func (r *PullRequestPostgresRepository) Create(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	prQuery := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, prQuery,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrPRExists
		}
		return err
	}

	if len(pr.AssignedReviewers) > 0 {
		reviewerQuery := `
			INSERT INTO pr_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)
		`

		for _, reviewerID := range pr.AssignedReviewers {
			_, err = tx.ExecContext(ctx, reviewerQuery, pr.PullRequestID, reviewerID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PullRequestPostgresRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	prQuery := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var pr domain.PullRequest
	var status string
	err := r.db.QueryRowContext(ctx, prQuery, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&status,
		&pr.CreatedAt,
		&pr.MergedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	pr.Status = domain.PRStatus(status)

	reviewersQuery := `
		SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY user_id
	`

	rows, err := r.db.QueryContext(ctx, reviewersQuery, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pr.AssignedReviewers = []string{}
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PullRequestPostgresRepository) GetByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	query := `
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var status string
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&status,
			&pr.CreatedAt,
			&pr.MergedAt,
		); err != nil {
			return nil, err
		}
		pr.Status = domain.PRStatus(status)

		reviewersQuery := `
			SELECT user_id
			FROM pr_reviewers
			WHERE pull_request_id = $1
			ORDER BY user_id
		`

		reviewerRows, err := r.db.QueryContext(ctx, reviewersQuery, pr.PullRequestID)
		if err != nil {
			return nil, err
		}

		pr.AssignedReviewers = []string{}
		for reviewerRows.Next() {
			var reviewerID string
			if err := reviewerRows.Scan(&reviewerID); err != nil {
				reviewerRows.Close()
				return nil, err
			}
			pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
		}
		reviewerRows.Close()

		if err := reviewerRows.Err(); err != nil {
			return nil, err
		}

		prs = append(prs, &pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

func (r *PullRequestPostgresRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	prQuery := `
		UPDATE pull_requests
		SET pull_request_name = $1, author_id = $2, status = $3, merged_at = $4
		WHERE pull_request_id = $5
	`

	result, err := tx.ExecContext(ctx, prQuery,
		pr.PullRequestName,
		pr.AuthorID,
		string(pr.Status),
		pr.MergedAt,
		pr.PullRequestID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrPRNotFound
	}

	deleteQuery := `DELETE FROM pr_reviewers WHERE pull_request_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, pr.PullRequestID)
	if err != nil {
		return err
	}

	if len(pr.AssignedReviewers) > 0 {
		reviewerQuery := `
			INSERT INTO pr_reviewers (pull_request_id, user_id)
			VALUES ($1, $2)
		`

		for _, reviewerID := range pr.AssignedReviewers {
			_, err = tx.ExecContext(ctx, reviewerQuery, pr.PullRequestID, reviewerID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PullRequestPostgresRepository) Exists(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, prID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
