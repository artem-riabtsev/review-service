package service

import (
	"context"
	"review-service/internal/model"
)

type Service interface {
	TeamService
	UserService
	PullRequestService
}

type TeamService interface {
	CreateTeam(ctx context.Context, team *model.Team) (*model.Team, error)
	GetTeam(ctx context.Context, teamName string) (*model.Team, error)
}

type UserService interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
	GetUserReviewRequests(ctx context.Context, userID string) ([]*model.PullRequestShort, error)
}

type PullRequestService interface {
	CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*model.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*model.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID string, oldUserID string) (*model.PullRequest, string, error)
}