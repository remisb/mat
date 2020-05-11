package restaurantapi

import (
	"github.com/go-chi/chi"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
)

func (s *Server) initRoutes() {
	if s.Router == nil {
		// /api/v1/restaurant
		s.Router = chi.NewMux()
		s.Router.Use(web.CorsHandler)

		s.Router.Get("/votes", s.handleMenuVotesGet)
		s.Router.Get("/menus", s.handleMenusGet)
		s.Router.Get("/", s.handleRestaurantsGet)
		s.Router.Get("/{restaurantId}", s.handleRestaurantGet)
		s.Router.Get("/{restaurantId}/menu", s.handleRestaurantMenusGet)
		s.Router.Get("/{restaurantId}/menu/:menuId", s.handleRestaurantMenuGet)

		s.Router.Group(func(r chi.Router) {
			r.Use(web.Verifier(s.jwtAuth))
			r.Use(web.Authenticator)

			r.Post("/{restaurantId}/menu/{menuId}/vote", s.handleRestaurantMenuVotePost)
			r.Post("/", s.handleRestaurantCreate)
			r.Put("/{restaurantId}", s.handleRestaurantUpdate())
			r.Delete("/{restaurantId}", s.handleRestaurantDelete())
			r.Post("/{restaurantId}/menu", s.handleRestaurantMenuCreate)
		})
	}
}
