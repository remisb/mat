package userapi

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/user"
	"net/http"
	"time"
)

func (s *Server) handlePolls(w http.ResponseWriter, r *http.Request) {
}

// Delete removes the specified user from the system.
func (s *Server) handleUserDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := chi.URLParam(r, "id")
		err := user.Delete(ctx, s.db, userID)
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
				web.RespondError(w, r, http.StatusInternalServerError, err)
			}
		}

		web.Respond(w, r, http.StatusNoContent, nil)
	}
}

func (s *Server) handleUserUpdate() http.HandlerFunc {
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

	uDb, err := user.Create(r.Context(), s.db, u.Name, u.Email, u.Password, u.Roles, time.Now())
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		fmt.Println("User created with id:", uDb.ID)
	}

	//w.Header().Set("Location", "/api/v1/users/" + uDb.ID)
	web.Respond(w, r, http.StatusCreated, uDb)
}

func (s *Server) handleToken2Get(w http.ResponseWriter, r *http.Request) {
	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		web.RespondError(w, r, http.StatusUnauthorized, err)
		return
	}

	ctx := r.Context()
	now := time.Now()
	claims, err := user.Authenticate(ctx, s.db, now, email, pass)
	if err != nil {
		switch err {
		case db.ErrAuthenticationFailure:
			web.RespondError(w, r, http.StatusUnauthorized, err)
			return
		default:
			err = errors.Wrap(err, "authenticating")
			web.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	token, tokenString, err := s.tokenAuth.Encode(claims)
	if err != nil {
		err = errors.Wrap(err, "token encode")
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	fmt.Printf("DEBUG: a sample jwt is %s\n \tclaim: %v\n", tokenString, token)

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token = tokenString
	web.Respond(w, r, http.StatusOK, tkn)
}

func (s *Server) handleTokenGet(w http.ResponseWriter, r *http.Request) {
	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		web.RespondError(w, r, http.StatusUnauthorized, err)
		return
	}

	ctx := r.Context()
	now := time.Now()
	claims, err := user.Authenticate(ctx, s.db, now, email, pass)
	if err != nil {
		switch err {
		case db.ErrAuthenticationFailure:
			web.RespondError(w, r, http.StatusUnauthorized, err)
			return
		default:
			err = errors.Wrap(err, "authenticating")
			web.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	token, tokenString, err := s.tokenAuth.Encode(claims)
	if err != nil {
		err = errors.Wrap(err, "token encode")
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	fmt.Printf("DEBUG: a sample jwt is %s\n \tclaim: %v\n", tokenString, token)

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token = tokenString
	web.Respond(w, r, http.StatusOK, tkn)
}

func (s *Server) handleTokenGetOriginal(w http.ResponseWriter, r *http.Request) {
	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		web.RespondError(w, r, http.StatusUnauthorized, err)
		return
	}

	ctx := r.Context()
	now := time.Now()
	claims, err := user.Authenticate(ctx, s.db, now, email, pass)
	if err != nil {
		switch err {
		case db.ErrAuthenticationFailure:
			web.RespondError(w, r, http.StatusUnauthorized, err)
			return
		default:
			err = errors.Wrap(err, "authenticating")
			web.RespondError(w, r, http.StatusInternalServerError, err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	_, tkn.Token, err = s.tokenAuth.Encode(claims)
	if err != nil {
		err = errors.Wrap(err, "generating token")
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	web.Respond(w, r, http.StatusOK, tkn)
}

func (s *Server) handleUsersPagedGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	users, err := user.List(r.Context(), s.db)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusOK, users)
}

func (s *Server) handleUsersGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	users, err := user.List(r.Context(), s.db)
	if err != nil {
		web.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}

	web.Respond(w, r, http.StatusOK, users)
}

func (s *Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		web.RespondError(w, r, http.StatusUnauthorized, web.ErrNoTokenFound)
		return
	}

	usr, err := user.Retrieve(ctx, claims, s.db, userID)
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
	web.Respond(w, r, http.StatusOK, usr)
}

func (s *Server) handleHealthGet(w http.ResponseWriter, r *http.Request) {
	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: s.build,
	}

	ctx := r.Context()
	// Check if the database is ready.
	if err := db.StatusCheck(ctx, s.db); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		web.Respond(w, r, http.StatusInternalServerError, health)
		return
	}

	health.Status = "ok"
	web.Respond(w, r, http.StatusOK, health)
}
