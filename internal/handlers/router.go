package handlers

import (
	"net/http"
	"review-service/internal/service"
)

type Router struct {
	userHandler *UserHandler
	teamHandler *TeamHandler
	prHandler   *PullRequestHandler
}

func NewRouter(
	userService service.UserService,
	teamService service.TeamService,
	prService service.PullRequestService,
) *Router {
	return &Router{
		userHandler: NewUserHandler(userService),
		teamHandler: NewTeamHandler(teamService),
		prHandler:   NewPullRequestHandler(prService),
	}
}

func (r *Router) SetupRoutes(mux *http.ServeMux) {

	mux.HandleFunc("/team/add", r.teamHandler.CreateTeam)
	mux.HandleFunc("/team/get", r.teamHandler.GetTeam)

	mux.HandleFunc("/health", HealthCheck)

	mux.HandleFunc("/users/setIsActive", r.userHandler.SetIsActive)
	mux.HandleFunc("/users/getReview", r.userHandler.GetReviews)

	mux.HandleFunc("/pullRequest/create", r.prHandler.CreatePullRequest)
	mux.HandleFunc("/pullRequest/merge", r.prHandler.MergePullRequest)
	mux.HandleFunc("/pullRequest/reassign", r.prHandler.ReassignReviewer)
}
