package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/auth"
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
		userId := chi.URLParam(r, "id")
		err := user.Delete(ctx, s.db, userId)
		if err != nil {
			switch err {
			case user.ErrInvalidID:
				err := web.NewRequestError(err, http.StatusBadRequest)
				s.respondError(w, r, http.StatusBadRequest, err)
				return
			case user.ErrNotFound:
				err := web.NewRequestError(err, http.StatusNotFound)
				s.respondError(w, r, http.StatusNotFound, err)
				return
			case user.ErrForbidden:
				err := web.NewRequestError(err, http.StatusForbidden)
				s.respondError(w, r, http.StatusForbidden, err)
				return
			default:
				err := errors.Wrapf(err, "Id: %s", userId)
				s.respondError(w, r, http.StatusInternalServerError, err)
			}
		}

		s.respond(w, r, http.StatusNoContent, nil)
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
		err := s.decodeBody(r, &updateUser)
		if err != nil {
			s.respondError(w, r, http.StatusBadRequest, errors.Wrap(err, ""))
			return
		}

		if err := decodeBody(r, &updateUser); err != nil {

		}
	}
}

func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	var u user.NewUser
	if err := decodeBody(r, &u); err != nil {
		respondErr(w, r, http.StatusBadRequest, "failed to read user from request", err)
		return
	}

	if u.Password != u.PasswordConfirm {
		respondErr(w, r, http.StatusBadRequest, "password and password confirm are not equal")
		return
	}

	//apiKey, ok := APIKey(r.Context())
	//if ok {
	//
	//}

	uDb, err := user.Create(r.Context(), s.db, u.Name, u.Email, u.Password, u.Roles, time.Now())
	if err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		fmt.Println("User created with id:", uDb.ID)
	}

	w.Header().Set("Location", "/api/v1/users/" + uDb.ID)
	s.respond(w, r, http.StatusCreated, nil)
}

func (s *Server) handleTokenGet(w http.ResponseWriter, r *http.Request) {
	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		s.respondError(w, r, http.StatusUnauthorized, err)
		return
	}

	ctx := r.Context()
	now := time.Now()
	claims, err := user.Authenticate(ctx, s.db, now, email, pass)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			s.respondError(w, r, http.StatusUnauthorized, err)
			return
		default:
			err = errors.Wrap(err, "authenticating")
			s.respondError(w, r, http.StatusInternalServerError, err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = s.Authenticator.GenerateToken(claims)
	if err != nil {
		err = errors.Wrap(err, "generating token")
		s.respondError(w, r, http.StatusInternalServerError, err)
		return
	}
	s.respond(w, r, http.StatusOK, tkn)
}

func (s *Server) handleUsersGet(w http.ResponseWriter, r *http.Request) {
	// TODO add pagination
	users, err := user.List(r.Context(), s.db)
	if err != nil {
		respondErr(w, r, http.StatusInternalServerError, err)
		return
	}

	respond(w, r, http.StatusOK, users)
}

func (s *Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := chi.URLParam(r, "id")
	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		err := errors.New("claims missing from context")
		s.respondError(w, r, http.StatusInternalServerError, err)
		return
	}

	usr, err := user.Retrieve(ctx, claims, s.db, userId)
	if err != nil {
		switch err {
		case user.ErrInvalidID:
			err := web.NewRequestError(err, http.StatusBadRequest)
			s.respondError(w, r, http.StatusBadRequest, err)
			return
		case user.ErrNotFound:
			err := web.NewRequestError(err, http.StatusNotFound)
			s.respondError(w, r, http.StatusNotFound, err)
			return
		case user.ErrForbidden:
			err := web.NewRequestError(err, http.StatusForbidden)
			s.respondError(w, r, http.StatusForbidden, err)
			return
		default:
			err := errors.Wrapf(err, "Id: %s", userId)
			s.respondError(w, r, http.StatusForbidden, err)
			return
		}
	}
	s.respond(w, r, http.StatusOK, usr)

}
