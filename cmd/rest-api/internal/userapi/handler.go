package userapi

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/pkg/errors"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/user"
	"net/http"
	"strings"
	"time"
)

func (s *Server) UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		ctx := r.Context()
		_, claims, err := jwtauth.FromContext(ctx)
		if err != nil || claims == nil {
			web.RespondError(w, r, http.StatusUnauthorized, web.ErrNoTokenFound)
			return
		}

		rolesSlice, err := rolesFromClaims(claims)

		isAdmin := hasRole(rolesSlice, auth.RoleAdmin)
		isUser := hasRole(rolesSlice, auth.RoleUser)

		user, err := s.userRepo.Retrieve(ctx, userID, rolesSlice)
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
				err := errors.Wrapf(err, "Id: %s", userID)
				web.RespondError(w, r, http.StatusForbidden, err)
				return
			}
		}

		if !isAdmin && isUser && user.Email != claims["email"] {
			err := errors.New("data is available only for owner")
			web.RespondError(w, r, http.StatusForbidden, err)
			return
		}

		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Delete removes the specified user from the system.
func (s *Server) handleUserDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		usr, ok := ctx.Value("user").(*user.User)
		if !ok {
			err := web.NewRequestError(web.ErrNoTokenFound, http.StatusBadRequest)
			web.RespondError(w, r, http.StatusNotFound, err)
			return
		}

		//userID := chi.URLParam(r, "id")
		err := s.userRepo.Delete(ctx, usr.ID)
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
				err := errors.Wrapf(err, "Id: %s", usr.ID)
				web.RespondError(w, r, http.StatusInternalServerError, err)
				return
			}
		}

		web.Respond(w, r, http.StatusOK, nil)
	}
}

func (s *Server) handleUserUpdate() http.HandlerFunc {
	type request struct {
		Name            *string  `json:"name"`
		Email           *string  `json:"email"`
		Roles           []string `json:"roles"`
		Password        *string  `json:"password"`
		PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {

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

func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	var u user.NewUser
	if err := web.DecodeBody(r, &u); err != nil {
		web.RespondError(w, r, http.StatusBadRequest, "failed to read user from request", err)
		return
	}

	if u.Password != u.PasswordConfirm {
		web.RespondError(w, r, http.StatusBadRequest, "password and password confirm are not equal")
		return
	}

	uDb, err := s.userRepo.Create(r.Context(), u.Name, u.Email, u.Password, u.Roles, time.Now())
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusCreated, uDb)
}

// handleTokenGet godoc
// @Summary Get JWT token
// @Description get jwt token
// @Accept  json
// @Produce  json
// @Success 200 {object} web.TokenResult
// @Failure 401 {object} web.APIError
// @Failure 500 {object} web.APIError
// @Router /users/token [get]
// @Security BasicAuth
func (s *Server) handleTokenGet(w http.ResponseWriter, r *http.Request) {
	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		web.RespondError(w, r, http.StatusUnauthorized, err)
		return
	}

	ctx := r.Context()

	aut := *s.authenticator
	token, _, err := aut.NewToken(ctx, email, pass)
	if err != nil {
		err = errors.Wrap(err, "token encode")
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	tkn := web.TokenResult{
		Token: token,
	}
	web.Respond(w, r, http.StatusOK, tkn)
}

// handleUsersGet godoc
// @Summary List users
// @Description get users
// @Accept   json
// @Produce  json
// @Success 200 {array} user.User
// @Failure 401 {object} web.APIError
// @Failure 403 {object} web.APIError
// @Failure 500 {object} web.APIError
// @Router /users [get]
func (s *Server) handleUsersGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil || claims == nil {
		web.RespondError(w, r, http.StatusUnauthorized, web.ErrNoTokenFound)
		return
	}

	rolesSlice, err := rolesFromClaims(claims)

	isAdmin := hasRole(rolesSlice, auth.RoleAdmin)

	if !isAdmin {
		err := errors.New("data is available only for admin")
		web.RespondError(w, r, http.StatusForbidden, err)
		return
	}

	users, err := s.userRepo.GetUsers(r.Context())
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusOK, users)
}

func (s *Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	usr, ok := r.Context().Value("user").(*user.User)
	if !ok {
		// in general this condition is not required, it should not happen
		err := errors.New("User not found")
		web.RespondError(w, r, http.StatusNotFound, err)
		return
	}
	web.Respond(w, r, http.StatusOK, usr)
}

func rolesStrToStrSlice(roles string) []string {
	return strings.Split(roles, ",")
}

func rolesFromClaims(claims jwt.MapClaims) ([]string, error) {
	roles2, ok := claims["roles"].([]interface{})
	if !ok {
		return nil, errors.New("invalid token, no roles")
	}

	rolesSlice := make([]string, len(roles2))
	for i, v := range roles2 {
		rolesSlice[i] = v.(string)
	}
	return rolesSlice, nil
}

func hasRole(roles []string, want string) bool {
	for _, role := range roles {
		if role == want {
			return true
		}
	}
	return false
}
