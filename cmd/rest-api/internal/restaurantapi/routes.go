package restaurantapi

import (
	"github.com/go-chi/chi"
	"github.com/remisb/mat/cmd/rest-api/internal/web"
)

func (s *Server) initRoutes() {
	if s.Router == nil {
		// /api/v1/restaurant
		restaurants := chi.NewMux()
		restaurants.Use(web.CorsHandler)

		restaurants.Get("/votes", s.handleMenuVotesGet)
		restaurants.Get("/menus", s.handleMenusGet)
		restaurants.Get("/", s.handleRestaurantsGet)
		restaurants.Get("/{restaurantId}", s.handleRestaurantGet)
		restaurants.Get("/{restaurantId}/menu", s.handleRestaurantMenusGet)
		restaurants.Get("/{restaurantId}/menu/:menuId", s.handleRestaurantMenuGet)

		restaurants.Group(func(r chi.Router) {
			r.Use(web.Verifier(s.jwtAuth))
			r.Use(web.Authenticator)

			r.Post("/{restaurantId}/menu/{menuId}/vote", s.handleRestaurantMenuVotePost)
			r.Post("/", s.handleRestaurantCreate)
			r.Put("/{restaurantId}", s.handleRestaurantUpdate())
			r.Delete("/{restaurantId}", s.handleRestaurantDelete())
			r.Post("/{restaurantId}/menu", s.handleRestaurantMenuCreate)
		})

		s.Router = restaurants
	}
}
