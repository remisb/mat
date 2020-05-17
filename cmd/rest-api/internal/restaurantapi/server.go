package restaurantapi

import (
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/restaurant"
	"github.com/remisb/mat/internal/user"
	"os"
)

// Server struct is a Restaurant REST API server
type Server struct {
	//Router http.Handler
	restaurantRepo *restaurant.Repo
	Router         *chi.Mux
	build          string
	authenticator  *auth.Authenticator
}

// NewServer is a factory function which creates and initializes new Restaurant REST API server.
func NewServer(build string, shutdown chan os.Signal, db *sqlx.DB) *Server {
	userRepo := user.NewRepo(db)
	s := Server{
		build:          build,
		authenticator:  auth.New(userRepo, web.Auth),
		restaurantRepo: restaurant.NewRepo(db),
	}

	s.initRoutes()
	return &s
}
