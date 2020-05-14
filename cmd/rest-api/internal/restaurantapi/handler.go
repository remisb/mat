package restaurantapi

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/pkg/errors"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/restaurant"
	"net/http"
	"time"
)

// handleRestaurantMenusGet is a handler function used to return list of restaurant menus of menus for specified restaurant
// handleRestaurantMenusGet godoc
// @Summary List of menus
// @Description get list of menus
// @Tags menus
// @Accept  json
// @Produce  json
// @param Authorization header string true "Authorization"
// @Param restaurantId path string true "Restaurant ID"
// @Param date query string false "name search by q" Format
// @Success 200 {array} restaurant.Menu
// @Failure 400 {object} web.APIError
// @Failure 404 {object} web.APIError
// @Failure 500 {object} web.APIError
// @Router /restaurant/{restaurantId}/menu [get]
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
	menus, err := s.restaurantRepo.RetrieveMenusByRestaurant(r.Context(), restaurantID)
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

	web.Respond(w, r, http.StatusOK, menus)
}

// handleRestaurantMenusCreate is a http handler function user to create
// handleRestaurantCreate godoc
// @Summary Add a restaurant menu
// @Description add new restaurant menu for current of specified date
// @Tags restaurants,menus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param restaurantId path string true "restaurant ID"
// @Param menu body restaurant.UpdateMenu true "update menu"
// @Success 200 {object} restaurant.Restaurant
// @Failure 400 {object} web.APIError
// @Failure 404 {object} web.APIError
// @Failure 500 {object} web.APIError
// @Router /restaurant/{restaurantId}/menu [post]
func (s *Server) handleRestaurantMenuCreate(w http.ResponseWriter, r *http.Request) {
	restaurantID := chi.URLParam(r, "restaurantId")
	if restaurantID == "" {
		web.RespondError(w, r, http.StatusBadRequest, "restaurantID is undefined")
		return
	}

	ctx := r.Context()
	restaurantItem, err := s.restaurantRepo.GetRestaurant(ctx, restaurantID)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	if restaurantItem == nil {
		web.RespondError(w, r, http.StatusNotFound)
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

	if claims == nil {
		web.RespondError(w, r, http.StatusUnauthorized, web.ErrNoTokenFound)
		return
	}
	//restaurant.UpdateMenu
	status := http.StatusOK
	if updateMenu.ID == "" {
		status = http.StatusCreated
	}

	menu, err := s.restaurantRepo.CreateRestaurantMenu(ctx, updateMenu)
	if err != nil {
		if err != restaurant.ErrNotFound {
			web.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	web.Respond(w, r, status, menu)
}

// endpoint: get /api/v1/restaurant/menus?date=2020-03-01
func (s *Server) handleMenusGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination

	parsedDate := parseURLDateDefaultNow(w, r, "date")
	todayMenus, err := s.restaurantRepo.RetrieveMenusByDate(r.Context(), parsedDate)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	web.Respond(w, r, http.StatusOK, todayMenus)
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

func (s *Server) handleRestaurantMenuGet(w http.ResponseWriter, r *http.Request) {
	restaurantID := chi.URLParam(r, "restaurantId")
	menuID := chi.URLParam(r, "menuId")

	menu, err := s.restaurantRepo.RetrieveRestaurantMenus(r.Context(), restaurantID, menuID)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, menu)
		return
	}

	web.Respond(w, r, http.StatusOK, menu)
}

// handleRestaurantsGet godoc
// @Summary List restaurant
// @Description get restaurants
// @Tags restaurants
// @Accept  json
// @Produce  json
// @Success 200 {array} restaurant.Restaurant
// @Failure 500 {object} web.APIError
// @Router /restaurant [get]
func (s *Server) handleRestaurantsGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	restaurants, err := s.restaurantRepo.RetrieveRestaurantList(r.Context())
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusOK, restaurants)
}

// handleRestaurantCreate godoc
// @Summary Add a restaurant
// @Description add new restaurant
// @Tags restaurants
// @Accept  json
// @Produce  json
// @Success 200 {object} restaurant.Restaurant
// @Failure 400 {object} web.APIError
// @Failure 500 {object} web.APIError
// @Router /restaurant [post]
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

	userID := claims["sub"].(string)

	uDb, err := s.restaurantRepo.CreateRestaurant(ctx, nr, time.Now(), userID)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusCreated, uDb)
}

// handleRestaurantsGet godoc
// @Summary List restaurant
// @Description get restaurants
// @ID get-restaurant-by-int
// @Tags restaurants
// @Accept  json
// @Produce  json
// @Param  restaurantId path string true "Restaurant ID"
// @Success 200 {array} restaurant.Restaurant
// @Failure 500 {object} web.APIError
// @Router /restaurant/{restaurantId} [get]
func (s *Server) handleRestaurantGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	restaurantID := chi.URLParam(r, "restaurantId")

	usr, err := s.restaurantRepo.GetRestaurant(ctx, restaurantID)
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

		err := s.restaurantRepo.DeleteRestaurant(ctx, restaurantID)
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
				return
			}
		}

		web.Respond(w, r, http.StatusOK, nil)
	}
}
