package repository

import (
	"context"
	"review-service/internal/model"
)

// TeamRepository интерфейс для работы с командами
type TeamRepository interface {
	CreateTeam(ctx context.Context, team *model.Team) error
	GetTeam(ctx context.Context, teamName string) (*model.Team, error)
	TeamExists(ctx context.Context, teamName string) (bool, error)
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, userID string) (*model.User, error)
	SetUserActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*model.User, error)
	UserExists(ctx context.Context, userID string) (bool, error)
}

// PullRequestRepository интерфейс для работы с PR
type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr *model.PullRequest, reviewers []string) error
	GetPullRequest(ctx context.Context, prID string) (*model.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) error
	ReassignReviewer(ctx context.Context, prID string, oldUserID string, newUserID string) error
	GetUserReviewRequests(ctx context.Context, userID string) ([]*model.PullRequestShort, error)
	PRExists(ctx context.Context, prID string) (bool, error)
	IsUserAssignedToPR(ctx context.Context, prID string, userID string) (bool, error)
}

// Объединяющий интерфейс
type Repository interface {
	TeamRepository
	UserRepository
	PullRequestRepository
}
