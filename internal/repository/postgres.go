package repository

import (
	"context"
	"review-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) CreateTeam(ctx context.Context, team *model.Team) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "INSERT INTO teams (team_name) VALUES ($1)", team.TeamName)
	if err != nil {
		return ErrTeamExists
	}

	for _, member := range team.Members {
		_, err = tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE SET
				username = EXCLUDED.username,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active,
				updated_at = NOW()
		`, member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) GetTeam(ctx context.Context, teamName string) (*model.Team, error) {
	var team model.Team
	team.TeamName = teamName

	rows, err := r.pool.Query(ctx, `
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var member model.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		team.Members = append(team.Members, member)
	}

	if len(team.Members) == 0 {
		return nil, ErrTeamNotFound
	}

	return &team, nil
}

func (r *postgresRepository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)", teamName).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) CreateUser(ctx context.Context, user *model.User) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
	`, user.UserID, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *postgresRepository) GetUser(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE user_id = $1
	`, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *postgresRepository) SetUserActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx, `
		UPDATE users 
		SET is_active = $1, updated_at = NOW() 
		WHERE user_id = $2 
		RETURNING user_id, username, team_name, is_active
	`, isActive, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	
	if err == pgx.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *postgresRepository) GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*model.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != $2
	`, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (r *postgresRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)", userID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) CreatePullRequest(ctx context.Context, pr *model.PullRequest, reviewers []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status) 
		VALUES ($1, $2, $3, $4)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, "OPEN")
	if err != nil {
		return ErrPRExists
	}

	for _, reviewerID := range reviewers {
		_, err = tx.Exec(ctx, `
			INSERT INTO pr_reviewers (pull_request_id, user_id) 
			VALUES ($1, $2)
		`, pr.PullRequestID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) GetPullRequest(ctx context.Context, prID string) (*model.PullRequest, error) {
	var pr model.PullRequest
	err := r.pool.QueryRow(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests 
		WHERE pull_request_id = $1
	`, prID).Scan(
		&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, 
		&pr.CreatedAt, &pr.MergedAt,
	)
	
	if err == pgx.ErrNoRows {
		return nil, ErrPRNotFound
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1
	`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return &pr, nil
}

func (r *postgresRepository) MergePullRequest(ctx context.Context, prID string) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE pull_requests 
		SET status = 'MERGED', merged_at = NOW(), updated_at = NOW()
		WHERE pull_request_id = $1 AND status = 'OPEN'
	`, prID)
	
	if err != nil {
		return err
	}
	
	if result.RowsAffected() == 0 {
		var exists bool
		err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return ErrPRNotFound
		}
	}
	
	return nil
}

func (r *postgresRepository) ReassignReviewer(ctx context.Context, prID string, oldUserID string, newUserID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM pull_requests WHERE pull_request_id = $1", prID).Scan(&status)
	if err == pgx.ErrNoRows {
		return ErrPRNotFound
	}
	if err != nil {
		return err
	}
	if status == "MERGED" {
		return ErrPRMerged
	}

	var assigned bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2)", prID, oldUserID).Scan(&assigned)
	if err != nil {
		return err
	}
	if !assigned {
		return ErrUserNotAssigned
	}

	_, err = tx.Exec(ctx, `
		UPDATE pr_reviewers 
		SET user_id = $1 
		WHERE pull_request_id = $2 AND user_id = $3
	`, newUserID, prID, oldUserID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) GetUserReviewRequests(ctx context.Context, userID string) ([]*model.PullRequestShort, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*model.PullRequestShort
	for rows.Next() {
		var pr model.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, &pr)
	}

	return prs, nil
}

func (r *postgresRepository) PRExists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID).Scan(&exists)
	return exists, err
}

func (r *postgresRepository) IsUserAssignedToPR(ctx context.Context, prID string, userID string) (bool, error) {
	var assigned bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM pr_reviewers 
			WHERE pull_request_id = $1 AND user_id = $2
		)
	`, prID, userID).Scan(&assigned)
	return assigned, err
}