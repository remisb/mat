package server

import (
	"net/http"
)

// InfoHandler is a HTTP handler function used to provide info about this microservice.
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Restaurant Menu Voting Service (c) 2020 Remis B"))
}
