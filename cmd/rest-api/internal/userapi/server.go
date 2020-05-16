package userapi

import (
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/user"
	"os"
)

// Server struct is a User REST API server
type Server struct {
	//Router http.Handler
	userRepo      *user.Repo
	Router        *chi.Mux
	build         string
	authenticator *auth.Authenticator
}

// NewServer is a factory function which creates and initializes new user REST API server.
func NewServer(build string, shutdown chan os.Signal, db *sqlx.DB) *Server {
	web.InitAuth()
	userRepo := user.NewRepo(db)
	s := Server{
		build:         build,
		authenticator: auth.New(userRepo, web.Auth),
		userRepo:      userRepo,
	}

	s.initRoutes()
	return &s
}
