package userapi

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"net/http"
	"os"
)

type Server struct {
	db *sqlx.DB
	//Router http.Handler
	Router        *chi.Mux
	build         string
	tokenAuth     *jwtauth.JWTAuth
}

func NewServer(build string, shutdown chan os.Signal, db *sqlx.DB) *Server {
	s := Server{
		db:            db,
		build:         build,
		tokenAuth:     web.TokenAuth,
	}

	s.routes()
	return &s
}

func (s Server) encodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return web.EncodeBody(w, r, v)
}

func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		web.EncodeBody(w, r, data)
	}
}
