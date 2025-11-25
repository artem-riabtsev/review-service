package handler

import (
	"encoding/json"
	"net/http"
	"review-service/internal/model"
	"review-service/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) http.Handler {
	h := &Handler{service: service}
	
	r := chi.NewRouter()
	
	// Teams endpoints
	r.Post("/team/add", h.createTeam)
	r.Get("/team/get", h.getTeam)
	
	// Users endpoints  
	r.Post("/users/setIsActive", h.setUserActive)
	r.Get("/users/getReview", h.getUserReviewRequests)
	
	// PullRequests endpoints
	r.Post("/pullRequest/create", h.createPullRequest)
	r.Post("/pullRequest/merge", h.mergePullRequest)
	r.Post("/pullRequest/reassign", h.reassignReviewer)
	
	// Health check
	r.Get("/health", h.healthCheck)
	
	return r
}

// Health check
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Teams handlers
func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	var req model.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	team, err := h.service.CreateTeam(r.Context(), &req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"team": team})
}

func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", "Missing team_name parameter"))
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, team)
}

// Users handlers
func (h *Handler) setUserActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	user, err := h.service.SetUserActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *Handler) getUserReviewRequests(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", "Missing user_id parameter"))
		return
	}

	prs, err := h.service.GetUserReviewRequests(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":        userID,
		"pull_requests": prs,
	})
}

// PullRequest handlers
func (h *Handler) createPullRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	pr, err := h.service.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"pr": pr})
}

func (h *Handler) mergePullRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	pr, err := h.service.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"pr": pr})
}

func (h *Handler) reassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", "Invalid request body"))
		return
	}

	pr, newUserID, err := h.service.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newUserID,
	})
}

// Вспомогательные функции
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, errorResp model.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResp)
}

func handleServiceError(w http.ResponseWriter, err error) {
	if businessErr, ok := err.(service.BusinessError); ok {
		switch businessErr.Code {
		case "TEAM_EXISTS":
			writeError(w, http.StatusBadRequest, model.NewErrorResponse("TEAM_EXISTS", businessErr.Message))
		case "PR_EXISTS":
			writeError(w, http.StatusConflict, model.NewErrorResponse("PR_EXISTS", businessErr.Message))
		case "PR_MERGED":
			writeError(w, http.StatusConflict, model.NewErrorResponse("PR_MERGED", businessErr.Message))
		case "NOT_ASSIGNED":
			writeError(w, http.StatusConflict, model.NewErrorResponse("NOT_ASSIGNED", businessErr.Message))
		case "NO_CANDIDATE":
			writeError(w, http.StatusConflict, model.NewErrorResponse("NO_CANDIDATE", businessErr.Message))
		case "NOT_FOUND":
			writeError(w, http.StatusNotFound, model.NewErrorResponse("NOT_FOUND", businessErr.Message))
		default:
			writeError(w, http.StatusBadRequest, model.NewErrorResponse("INVALID_INPUT", businessErr.Message))
		}
		return
	}
	
	// Общая ошибка сервера
	writeError(w, http.StatusInternalServerError, model.NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
}