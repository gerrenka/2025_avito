package handlers

import (
	"encoding/json"
	"net/http"
	"review-service/internal/domain"
	"review-service/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, domain.ErrCodeNotFound, "method not allowed")
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "invalid request body")
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"user": user,
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, domain.ErrCodeNotFound, "method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "user_id is required")
		return
	}

	prs, err := h.userService.GetReviews(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	}
	writeJSON(w, http.StatusOK, response)
}
