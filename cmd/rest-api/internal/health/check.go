package health

import (
	"github.com/remisb/mat/cmd/rest-api/internal/web"
	"github.com/remisb/mat/internal/db"
	"net/http"
)

// Health validates the service is healthy and ready to accept requests.
func handleHealthGet(w http.ResponseWriter, r *http.Request) {

	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: conf.build,
	}

	// Check if the database is ready.
	if err := db.StatusCheck(r.Context(), conf.db); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		web.Respond(w, r, http.StatusInternalServerError, health)
		return
	}

	health.Status = "ok"
	web.Respond(w, r, http.StatusOK, health)
}
