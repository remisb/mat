package server

import (
	"net/http"
)

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Restaurant Menu Voting Service (c) 2020 Remis B"))
}
