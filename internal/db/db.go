package db

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // The database driver in use.
	"net/url"
)

var (
	// ErrNotFound used when user is not found
	ErrNotFound = errors.New("User not found")
	// ErrInvalidID used when passed ID has invalid format
	ErrInvalidID = errors.New("ID is not in its proper form")
	// ErrAuthenticationFailure used when authentication has failed
	ErrAuthenticationFailure = errors.New("AuthenticationFailed")
	// ErrForbidden used when forbitten action was tryed to perform.
	ErrForbidden = errors.New("Attempted action is not allowed")
	// ErrAlreadyVoted used when user is trying to place second vote for the same date.
	ErrAlreadyVoted = errors.New("user has already voted today")
)

type Config struct {
	Host       string
	Port       string
	User       string
	Password   string
	Name       string
	DisableTLS bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*sqlx.DB, error) {

	// Define SSL mode.
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	// Query parameters.
	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	// Construct url.
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	return sqlx.Open("postgres", u.String())
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	// Run a simple query to determine connectivity. The db has a "Ping" method
	// but it can false-positive when it was previously able to talk to the
	// database but the database has since gone away. Running this query forces a
	// round trip to the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}
