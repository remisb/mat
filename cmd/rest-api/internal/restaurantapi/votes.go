package restaurantapi

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/db"
	"net/http"
)

// endpoint: GET /api/v1/restaurant/votes?date=2020-03-02
func (s *Server) handleMenuVotesGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	parsedDate := parseURLDateDefaultNow(w, r, "date")
	menuVotes, err := s.restaurantRepo.MenuVotes(r.Context(), parsedDate)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusOK, menuVotes)
}

// handleRestaurantMenuVotePost is used to vote.
//
// vote is allowed only for registered user
// one menu vote is allowed per day
// removal / change or update of vote is not allowed
// vote is allowed only for today's menu
//
// endpoint: POST /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
//
func (s *Server) handleRestaurantMenuVotePost(w http.ResponseWriter, r *http.Request) {
	restaurantID := chi.URLParam(r, "restaurantId")
	menuID := chi.URLParam(r, "menuId")

	ctx := r.Context()

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		web.RespondError(w, r, http.StatusUnauthorized, web.ErrNoTokenFound)
		return
	}
	userID := claims["sub"].(string)

	parsedDate := parseURLDateDefaultNow(w, r, "date")

	err = s.restaurantRepo.MenuVote(ctx, userID, restaurantID, menuID, parsedDate)
	if err != nil {
		if err == db.ErrAlreadyVoted {
			web.RespondError(w, r, http.StatusForbidden, err)
			return
		}
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	response := map[string]string{
		"success": "vote accepted",
	}

	web.Respond(w, r, http.StatusCreated, response)
}
