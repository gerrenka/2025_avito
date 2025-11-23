package handlers

import (
	"encoding/json"
	"net/http"

	"review-service/internal/domain"
	"review-service/internal/service"
)

type PullRequestHandler struct {
	prService service.PullRequestService
}

func NewPullRequestHandler(prService service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		prService: prService,
	}
}

func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, domain.ErrCodeNotFound, "method not allowed")
		return
	}

	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "missing required fields")
		return
	}

	pr, err := h.prService.CreatePullRequest(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"pr": pr,
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *PullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, domain.ErrCodeNotFound, "method not allowed")
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "invalid request body")
		return
	}

	if req.PullRequestID == "" {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "pull_request_id is required")
		return
	}

	pr, err := h.prService.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"pr": pr,
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *PullRequestHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, domain.ErrCodeNotFound, "method not allowed")
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "missing required fields")
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	}
	writeJSON(w, http.StatusOK, response)
}
