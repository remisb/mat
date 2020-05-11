package userapi

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/user"
	"os"
)

type Server struct {
	//Router http.Handler
	db            *sqlx.DB
	userRepo      *user.Repo
	Router        *chi.Mux
	build         string
	jwtAuth       *jwtauth.JWTAuth
	authenticator *auth.Authenticator
}

func NewServer(build string, shutdown chan os.Signal, db *sqlx.DB) *Server {
	userRepo := user.NewRepo(db)
	jwtauth := jwtauth.New("HS256", []byte("secret"), nil)
	s := Server{
		db:            db,
		build:         build,
		jwtAuth:       jwtauth,
		authenticator: auth.New(userRepo, jwtauth),
		userRepo:      userRepo,
	}

	s.initRoutes()
	return &s
}
