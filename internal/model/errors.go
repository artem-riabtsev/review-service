package model

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

const (
	ErrorTeamExists    = "TEAM_EXISTS"
	ErrorPRExists      = "PR_EXISTS"
	ErrorPRMerged      = "PR_MERGED"
	ErrorNotAssigned   = "NOT_ASSIGNED"
	ErrorNoCandidate   = "NO_CANDIDATE"
	ErrorNotFound      = "NOT_FOUND"
)

func NewErrorResponse(code, message string) ErrorResponse {
	var resp ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	return resp
}