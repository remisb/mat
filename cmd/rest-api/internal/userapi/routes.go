package userapi

import (
	"github.com/go-chi/chi"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
)

func (s *Server) initRoutes() {
	if s.Router == nil {
		users := chi.NewMux()
		users.Use(web.CorsHandler)

		// /api/v1/users/
		users.Group(func(r chi.Router) {
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

		users.Get("/token", s.handleTokenGet)
		users.Get("/health", s.handleHealthGet)

		s.Router = users
	}
}
