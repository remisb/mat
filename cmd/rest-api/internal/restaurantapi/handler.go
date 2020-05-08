package restaurantapi

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/pkg/errors"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/restaurant"
	"net/http"
	"time"
)

// RetrieveRestaurantList of menus for specified restaurant
// endpoint: /api/v1/restaurant/:restaurantId/menu
func (s *Server) handleRestaurantMenusGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	// TODO menu list should be accessible to anyone
	// TODO restaurant menus should be ordered from the latest to the oldest in descendinf order
	//      based on the menu date

	restaurantID := chi.URLParam(r, "restaurantId")
	if restaurantID == "" {
		web.RespondError(w, r, http.StatusBadRequest, "restaurantID is undefined")
		return
	}
	restaurants, err := restaurant.RetrieveMenusByRestaurant(r.Context(), s.db, restaurantID)
	if err != nil {
		switch err {
		case db.ErrInvalidID:
			err := web.NewRequestError(err, http.StatusBadRequest)
			web.RespondError(w, r, http.StatusBadRequest, err)
			return
		case restaurant.ErrRestaurantNotFound:
			err := web.NewRequestError(err, http.StatusNotFound)
			web.RespondError(w, r, http.StatusNotFound, err)
			return
		case db.ErrForbidden:
			err := web.NewRequestError(err, http.StatusForbidden)
			web.RespondError(w, r, http.StatusForbidden, err)
			return
		default:
			err := errors.Wrapf(err, "Id: %s", restaurantID)
			web.RespondError(w, r, http.StatusForbidden, err)
			return
		}
	}

	web.Respond(w, r, http.StatusOK, restaurants)
}

// handleRestaurantMenusCreate is a http handler function user to create
// or update restaurant menu for specified date.
// endpoint: POST /api/v1/restaurant/{restaurantId}/menu
func (s *Server) handleRestaurantMenuCreate(w http.ResponseWriter, r *http.Request) {
	restaurantID := chi.URLParam(r, "restaurantId")
	if restaurantID == "" {
		web.RespondError(w, r, http.StatusBadRequest, "restaurantID is undefined")
		return
	}

	ctx := r.Context()
	restaurantItem, err := restaurant.RetrieveRestaurant(ctx, s.db, restaurantID)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	if restaurantItem == nil {
		web.RespondError(w,r,http.StatusNotFound)
		return
	}

	var updateMenu restaurant.UpdateMenu
	if err := web.DecodeBody(r, &updateMenu); err != nil {
		web.RespondError(w, r, http.StatusBadRequest, "failed to read updateMenu from request ", err)
		return
	}

	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		web.RespondError(w, r, http.StatusUnauthorized, web.ErrNoTokenFound)
		return
	}

	//restaurant.UpdateMenu
	status := http.StatusOK
	if updateMenu.ID == "" {
		status = http.StatusCreated
	}

	menu, err := restaurant.CreateRestaurantMenu(ctx, claims, s.db, updateMenu)
	if err != nil {
		if err != restaurant.ErrNotFound {
			web.RespondError(w, r, http.StatusInternalServerError, err)
		}
	}

	web.Respond(w,r, status, menu)
}

// endpoint: get /api/v1/restaurant/menus?date=2020-03-01
func (s *Server) handleMenusGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination

	parsedDate := parseURLDateDefaultNow(w, r, "date")
	todayMenus, err := restaurant.RetrieveMenusByDate(r.Context(), s.db, parsedDate)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	web.Respond(w, r, http.StatusOK, todayMenus)
}

// endpoint: GET /api/v1/restaurant/votes?date=2020-03-02
func (s *Server) handleMenuVotesGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	parsedDate := parseURLDateDefaultNow(w, r, "date")
	menuVotes, err := restaurant.MenuVotes(r.Context(), s.db, parsedDate)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusOK, menuVotes)
}

func parseURLDateDefaultNow(w http.ResponseWriter, r *http.Request, name string) time.Time {
	var parsedDate time.Time
	var err error

	date := r.URL.Query().Get(name)
	if date != "" {
		parsedDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			err = errors.Wrap(err, "invalid date format")
			web.RespondError(w, r, http.StatusBadRequest)
			return time.Now()
		}
	}

	if parsedDate.IsZero() {
		parsedDate = time.Now()
	}
	return parsedDate
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

	parsedDate := parseURLDateDefaultNow(w, r, "date")
	err = restaurant.MenuVote(ctx, claims, s.db, restaurantID, menuID, parsedDate)
	if err != nil {
		if err.Error() == "user has already voted today" {
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

func (s *Server) handleRestaurantMenuGet(w http.ResponseWriter, r *http.Request) {
	restaurantID := chi.URLParam(r, "restaurantId")
	menuID := chi.URLParam(r, "menuId")

	menu, err := restaurant.RetrieveRestaurantMenus(r.Context(), s.db, restaurantID, menuID)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, menu)
		return
	}

	web.Respond(w, r, http.StatusOK, menu)
}

// handlerFunc
// endpoint: GET /api/v1/restaurant
func (s *Server) handleRestaurantsGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	restaurants, err := restaurant.RetrieveRestaurantList(r.Context(), s.db)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusOK, restaurants)
}
// endpoint: POST /api/v1/restaurant
func (s *Server) handleRestaurantCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		err := errors.New("error on getting claims from context")
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	var nr restaurant.NewRestaurant
	if err := web.DecodeBody(r, &nr); err != nil {
		web.RespondError(w, r, http.StatusBadRequest, "failed to read user from request", err)
		return
	}

	if len(nr.Name) == 0 {
		web.RespondError(w, r, http.StatusBadRequest, "name of the restaurant should not be empty.")
		return
	}

	uDb, err := restaurant.CreateRestaurant(ctx, claims, s.db, nr, time.Now())
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		fmt.Println("User created with id:", uDb.ID)
	}

	//w.Header().Set("Location", "/api/v1/users/" + uDb.ID)
	web.Respond(w, r, http.StatusCreated, uDb)
}

func (s *Server) handleRestaurantGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	restaurantID := chi.URLParam(r, "restaurantId")

	usr, err := restaurant.RetrieveRestaurant(ctx, s.db, restaurantID)
	if err != nil {
		switch err {
		case db.ErrInvalidID:
			err := web.NewRequestError(err, http.StatusBadRequest)
			web.RespondError(w, r, http.StatusBadRequest, err)
			return
		case restaurant.ErrRestaurantNotFound:
			err := web.NewRequestError(err, http.StatusNotFound)
			web.RespondError(w, r, http.StatusNotFound, err)
			return
		case db.ErrForbidden:
			err := web.NewRequestError(err, http.StatusForbidden)
			web.RespondError(w, r, http.StatusForbidden, err)
			return
		default:
			err := errors.Wrapf(err, "Id: %s", restaurantID)
			web.RespondError(w, r, http.StatusForbidden, err)
			return
		}
	}
	web.Respond(w, r, http.StatusOK, usr)
}

func (s *Server) handleRestaurantUpdate() http.HandlerFunc {
	// UpdateUser defines what information may be provided to modify an existing
	// User. All fields are optional so clients can send just the fields they want
	// changed. It uses pointer fields so we can differentiate between a field that
	// was not provided and a field that was provided as explicitly blank. Normally
	// we do not want to use pointers to basic types but we make exceptions around
	// marshalling/unmarshalling.
	type request struct {
		Name            *string  `json:"name"`
		Email           *string  `json:"email"`
		Roles           []string `json:"roles"`
		Password        *string  `json:"password"`
		PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		restaurantID := chi.URLParam(r, "restaurantId")
		if restaurantID == "" {
			web.RespondError(w, r, http.StatusBadRequest, "restaurantID is undefined")
			return
		}

		var updateUser request
		err := web.DecodeBody(r, &updateUser)
		if err != nil {
			web.RespondError(w, r, http.StatusBadRequest, errors.Wrap(err, ""))
			return
		}

		if err := web.DecodeBody(r, &updateUser); err != nil {

		}
	}
}

// DeleteRestaurant removes the specified user from the system.
func (s *Server) handleRestaurantDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		restaurantID := chi.URLParam(r, "restaurantId")
		if restaurantID == "" {
			web.RespondError(w, r, http.StatusBadRequest, "restaurantID is undefined")
			return
		}

		err := restaurant.DeleteRestaurant(ctx, s.db, restaurantID)
		if err != nil {
			switch err {
			case db.ErrInvalidID:
				err := web.NewRequestError(err, http.StatusBadRequest)
				web.RespondError(w, r, http.StatusBadRequest, err)
				return
			case db.ErrNotFound:
				err := web.NewRequestError(err, http.StatusNotFound)
				web.RespondError(w, r, http.StatusNotFound, err)
				return
			case db.ErrForbidden:
				err := web.NewRequestError(err, http.StatusForbidden)
				web.RespondError(w, r, http.StatusForbidden, err)
				return
			default:
				err := errors.Wrapf(err, "Id: %s", restaurantID)
				web.RespondError(w, r, http.StatusInternalServerError, err)
			}
		}

		web.Respond(w, r, http.StatusNoContent, nil)
	}
}
