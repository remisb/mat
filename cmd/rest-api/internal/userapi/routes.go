package userapi

import (
	"github.com/go-chi/chi"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
)

func (s *Server) initRoutes() {
	if s.Router == nil {
		s.Router = chi.NewMux()
		s.Router.Use(web.CorsHandler)

		// /api/v1/users/
		s.Router.Group(func(r chi.Router) {
			r.Use(web.Verifier(s.jwtAuth))
			r.Use(web.Authenticator)

			r.Get("/", s.handleUsersGet)
			r.Post("/", s.handleUserCreate)
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(s.UserCtx)
				r.Get("/", s.handleUserGet)
				r.Put("/", s.handleUserUpdate())
				r.Delete("/", s.handleUserDelete())
			})
		})
		s.Router.Get("/token", s.handleTokenGet)
		s.Router.Get("/api/v1/health", s.handleHealthGet)
	}
}
