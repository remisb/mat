package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/internal/auth"
	"net/http"
)

type Server struct {
	db     *sqlx.DB
	//Router http.Handler
	Router *chi.Mux
	Authenticator *auth.Authenticator
}

func (s Server) decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func (s Server) respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		encodeBody(w, r, data)
	}
}

func (s Server) respondError(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	s.respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...),},
	})
}
