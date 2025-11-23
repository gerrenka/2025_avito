package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"review-service/internal/domain"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code domain.ErrorCode, message string) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    string(code),
			Message: message,
		},
	}
	writeJSON(w, status, response)
}

func handleServiceError(w http.ResponseWriter, err error) {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		status := getHTTPStatus(domainErr.Code)
		writeError(w, status, domainErr.Code, domainErr.Message)
		return
	}

	writeError(w, http.StatusInternalServerError, domain.ErrCodeNotFound, "internal server error")
}

func getHTTPStatus(code domain.ErrorCode) int {
	switch code {
	case domain.ErrCodeTeamExists:
		return http.StatusBadRequest // 400
	case domain.ErrCodePRExists:
		return http.StatusConflict // 409
	case domain.ErrCodeNotFound:
		return http.StatusNotFound // 404
	case domain.ErrCodePRMerged, domain.ErrCodeNotAssigned, domain.ErrCodeNoCandidate:
		return http.StatusConflict // 409
	default:
		return http.StatusInternalServerError // 500
	}
}
