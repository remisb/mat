package health

import (
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"net/http"
)

type config struct {
	build string
	db    *sqlx.DB
}

var conf config

func InitRouter(build string, db *sqlx.DB) http.Handler {
	conf.build = build
	conf.db = db

	r := chi.NewRouter()
	r.Get("/health", handleHealthGet)
	return r
}
