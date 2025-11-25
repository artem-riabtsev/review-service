package service

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrNoReviewerCandidate = errors.New("no reviewer candidate available")
	ErrAuthorNotFound   = errors.New("author not found")
	ErrTeamNotFound     = errors.New("team not found")
	ErrUserNotInTeam    = errors.New("user not in team")
)

type BusinessError struct {
	Code    string
	Message string
	Err     error
}

func (e BusinessError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func NewBusinessError(code, message string, err error) BusinessError {
	return BusinessError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}