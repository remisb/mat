package userapi

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
)

func (s *Server) routes() {
	if s.Router == nil {
		s.Router = chi.NewMux()
		s.Router.Use(web.CorsHandler)

		// /api/v1/users/
		s.Router.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(s.tokenAuth))

			// Handle valid / invalid tokens. In this example, we use
			// the provided authenticator middleware, but you can write your
			// own very easily, look at the Authenticator method in jwtauth.go
			// and tweak it, its not scary.
			r.Use(jwtauth.Authenticator)

			r.Get("/", s.handleUsersGet)
			r.Post("/", s.handleUserCreate)
			r.Get("/{id}", s.handleUserGet)
			r.Put("/{id}", s.handleUserUpdate())
			r.Delete("/{id}", s.handleUserDelete())
		})
		s.Router.Get("/token", s.handleTokenGet)
		s.Router.Get("/api/v1/health", s.handleHealthGet)
	}
}
