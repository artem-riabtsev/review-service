package service

import (
	"context"
	"math/rand"
	"review-service/internal/model"
	"review-service/internal/repository"
	"time"
)

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateTeam(ctx context.Context, team *model.Team) (*model.Team, error) {
	if team.TeamName == "" || len(team.Members) == 0 {
		return nil, ErrInvalidInput
	}

	exists, err := s.repo.TeamExists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, NewBusinessError("TEAM_EXISTS", "team already exists", nil)
	}

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *service) GetTeam(ctx context.Context, teamName string) (*model.Team, error) {
	if teamName == "" {
		return nil, ErrInvalidInput
	}

	team, err := s.repo.GetTeam(ctx, teamName)
	if err != nil {
		if err == repository.ErrTeamNotFound {
			return nil, NewBusinessError("NOT_FOUND", "team not found", err)
		}
		return nil, err
	}

	return team, nil
}

func (s *service) SetUserActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.SetUserActive(ctx, userID, isActive)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, NewBusinessError("NOT_FOUND", "user not found", err)
		}
		return nil, err
	}

	return user, nil
}

func (s *service) GetUserReviewRequests(ctx context.Context, userID string) ([]*model.PullRequestShort, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}

	exists, err := s.repo.UserExists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, NewBusinessError("NOT_FOUND", "user not found", nil)
	}

	return s.repo.GetUserReviewRequests(ctx, userID)
}

func (s *service) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*model.PullRequest, error) {
	if prID == "" || prName == "" || authorID == "" {
		return nil, ErrInvalidInput
	}

	author, err := s.repo.GetUser(ctx, authorID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, NewBusinessError("NOT_FOUND", "author not found", err)
		}
		return nil, err
	}

	team, err := s.repo.GetTeam(ctx, author.TeamName)
	if err != nil {
		if err == repository.ErrTeamNotFound {
			return nil, NewBusinessError("NOT_FOUND", "team not found", err)
		}
		return nil, err
	}

	reviewers := s.selectReviewers(team.Members, authorID)

	now := time.Now()
	pr := &model.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  prName,
		AuthorID:         authorID,
		Status:           "OPEN",
		AssignedReviewers: reviewers,
		CreatedAt:        &now,
	}

	if err := s.repo.CreatePullRequest(ctx, pr, reviewers); err != nil {
		if err == repository.ErrPRExists {
			return nil, NewBusinessError("PR_EXISTS", "PR id already exists", err)
		}
		return nil, err
	}

	return pr, nil
}

func (s *service) MergePullRequest(ctx context.Context, prID string) (*model.PullRequest, error) {
	if prID == "" {
		return nil, ErrInvalidInput
	}

	_, err := s.repo.GetPullRequest(ctx, prID)
	if err != nil {
		if err == repository.ErrPRNotFound {
			return nil, NewBusinessError("NOT_FOUND", "PR not found", err)
		}
		return nil, err
	}

	if err := s.repo.MergePullRequest(ctx, prID); err != nil {
		return nil, err
	}

	return s.repo.GetPullRequest(ctx, prID)
}

func (s *service) ReassignReviewer(ctx context.Context, prID string, oldUserID string) (*model.PullRequest, string, error) {
	if prID == "" || oldUserID == "" {
		return nil, "", ErrInvalidInput
	}

	pr, err := s.repo.GetPullRequest(ctx, prID)
	if err != nil {
		if err == repository.ErrPRNotFound {
			return nil, "", NewBusinessError("NOT_FOUND", "PR not found", err)
		}
		return nil, "", err
	}

	if pr.Status == "MERGED" {
		return nil, "", NewBusinessError("PR_MERGED", "cannot reassign on merged PR", nil)
	}

	assigned, err := s.repo.IsUserAssignedToPR(ctx, prID, oldUserID)
	if err != nil {
		return nil, "", err
	}
	if !assigned {
		return nil, "", NewBusinessError("NOT_ASSIGNED", "reviewer is not assigned to this PR", nil)
	}

	oldReviewer, err := s.repo.GetUser(ctx, oldUserID)
	if err != nil {
		return nil, "", err
	}

	newReviewerID, err := s.selectReplacementReviewer(ctx, oldReviewer.TeamName, oldUserID, pr.AssignedReviewers)
	if err != nil {
		return nil, "", NewBusinessError("NO_CANDIDATE", "no active replacement candidate in team", err)
	}

	if err := s.repo.ReassignReviewer(ctx, prID, oldUserID, newReviewerID); err != nil {
		return nil, "", err
	}

	updatedPR, err := s.repo.GetPullRequest(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewerID, nil
}

func (s *service) selectReviewers(members []model.TeamMember, excludeUserID string) []string {
	var candidateIDs []string
	
	for _, member := range members {
		if member.IsActive && member.UserID != excludeUserID {
			candidateIDs = append(candidateIDs, member.UserID)
		}
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(candidateIDs), func(i, j int) {
		candidateIDs[i], candidateIDs[j] = candidateIDs[j], candidateIDs[i]
	})

	if len(candidateIDs) > 2 {
		return candidateIDs[:2]
	}
	return candidateIDs
}

func (s *service) selectReplacementReviewer(ctx context.Context, teamName string, excludeUserID string, currentReviewers []string) (string, error) {
	candidates, err := s.repo.GetActiveUsersByTeam(ctx, teamName, excludeUserID)
	if err != nil {
		return "", err
	}

	var availableCandidates []string
	for _, candidate := range candidates {
		if !contains(currentReviewers, candidate.UserID) {
			availableCandidates = append(availableCandidates, candidate.UserID)
		}
	}

	if len(availableCandidates) == 0 {
		return "", ErrNoReviewerCandidate
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	return availableCandidates[rand.Intn(len(availableCandidates))], nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}