package handlers

import (
	"encoding/json"
	"net/http"
	"review-service/internal/domain"
	"review-service/internal/service"
)

type TeamHandler struct {
	teamService service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, domain.ErrCodeNotFound, "method not allowed")
		return
	}

	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "invalid request body")
		return
	}

	createdTeam, err := h.teamService.CreateTeam(r.Context(), &team)
	if err != nil {

		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"team": createdTeam,
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, domain.ErrCodeNotFound, "method not allowed")
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, http.StatusBadRequest, domain.ErrCodeNotFound, "team_name is required")
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, team)
}
