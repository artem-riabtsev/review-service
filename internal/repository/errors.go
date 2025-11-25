package repository

import "errors"

var (
	ErrTeamExists      = errors.New("team already exists")
	ErrTeamNotFound    = errors.New("team not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrPRNotFound      = errors.New("pull request not found")
	ErrPRExists        = errors.New("pull request already exists")
	ErrPRMerged        = errors.New("pull request is merged")
	ErrUserNotAssigned = errors.New("user is not assigned as reviewer")
	ErrNoActiveUsers   = errors.New("no active users available")
)