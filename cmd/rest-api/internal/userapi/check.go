package userapi

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/internal/db"
	"net/http"
)

// Check provides support for orchestration health checks.
type Check struct {
	build string
	db    *sqlx.DB
}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) {

	health := struct {
		Version string `json:"version"`
		Status  string `json:"status"`
	}{
		Version: c.build,
	}

	// Check if the database is ready.
	if err := db.StatusCheck(ctx, c.db); err != nil {

		// If the database is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.
		health.Status = "db not ready"
		Respond(w, r, http.StatusInternalServerError, health)
	}

	health.Status = "ok"
	Respond(w, r, http.StatusOK, health)
}
