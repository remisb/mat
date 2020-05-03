package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func initRoutes(s *Server) {
	if s.Router == nil {
		s.Router = chi.NewMux()
		s.Router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}))


		s.Router.Group(func(r chi.Router) {
			r.Get("/v1/users", s.handleUsersGet)
			r.Post("/v1/users", s.handleUserCreate)
			r.Get("/v1/users/:id", s.handleUserGet)
			r.Put("/v1/users/:id", s.handleUserUpdate())
			r.Delete("/v1/users/:id", s.handleUserDelete())
		})

		s.Router.Get("/v1/user/token", s.handleTokenGet)

		//mux.HandleFunc("/polls/", withCORS(withAPIKey(s.handlePolls)))
		//mux.HandleFunc("/api/v1/users", withCORS(withAPIKey(s.handleUsersGet)))
		//mux.HandleFunc("/api/v1/users", withCORS(withAPIKey(s.handleUserPost)))
		//mux.HandleFunc("/api/v1/users/", withCORS(withAPIKey(s.handleUsersGet)))
		//mux.HandleFunc("/api/v1/users/:userId", withCORS(withAPIKey(s.handleUserDelete)))

		// r.Get("/v1/users", s.handleUsersGet())
		// r.Post("/v1/users", s.handleUserCreate())
		// r.Get("/v1/users/:id", s.handleUserGet())
		// r.Put("/v1/users/:id", s.handleUserUpdate())
		// r.Delete("/v1/users/:id", s.handleUserDelete())
	}
}
