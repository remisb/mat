package restaurantapi

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
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
