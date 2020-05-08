package tests

import (
	"context"
	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/db/dbtest"
	logz "github.com/remisb/mat/internal/log"
	"github.com/remisb/mat/internal/user"
	"go.uber.org/zap"

	"github.com/remisb/mat/internal/schema"
	"testing"
	"time"
)

// Test structure is used to perform Integration tests.
type Test struct {
	DB         *sqlx.DB
	Log        *zap.SugaredLogger
	t          *testing.T
	cleanup    func()
	tokenAuth  *jwtauth.JWTAuth
	UserToken  string
	User1Token string
	User2Token string
	AdminToken string
}

func NewUnit(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	c := dbtest.StartContainer(t)
	db, err := db.Open(db.Config{
		User:       "postgres",
		Password:   "postgres",
		Host:       c.Host,
		Name:       "postgres",
		DisableTLS: true,
	})
	if err != nil {
		t.Fatalf("opening database connection: %v", err)
	}

	t.Log("waiting for database to be ready")

	// Wait for the database to be ready. Wait 100ms longer between each attempt.
	// Do not try more than 20 times.
	var pingError error
	maxAttempts := 20
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(maxAttempts) * 100 * time.Millisecond)
	}

	if pingError != nil {
		dbtest.DumpContainerLogs(t, c)
		dbtest.StopContainer(t, c)
		t.Fatalf("waiting for database to be ready: %v", pingError)
	}

	if err := schema.Migrate(db); err != nil {
		dbtest.StopContainer(t, c)
		t.Fatalf("migrating: %s", err)
	}

	teardown := func() {
		t.Helper()
		db.Close()
		dbtest.StopContainer(t, c)
	}

	return db, teardown
}

// NewIntegration provides a Test struct setup for integration testing.
func NewIntegration(t *testing.T) *Test {
	t.Helper()

	db, cleanup := NewUnit(t)
	if err := schema.Seed(db); err != nil {
		t.Fatal(err)
	}

	// Create the logger to use.
	//logger := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Build an authenticator using this static key.

	//tokenAuth := web.TokenAuth
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	return &Test{
		DB:         db,
		Log:        logz.Sugar,
		t:          t,
		cleanup:    cleanup,
		UserToken:  Token(t, db, tokenAuth, "user@example.com", "gophers"),
		User1Token: Token(t, db, tokenAuth, "user1@example.com", "gophers"),
		User2Token: Token(t, db, tokenAuth, "user2@example.com", "gophers"),
		AdminToken: Token(t, db, tokenAuth, "admin@example.com", "gophers"),
		tokenAuth:  tokenAuth,
	}
}

// Token generates an authenticated token for a user.
func Token(t *testing.T, DB *sqlx.DB, tokenAuth *jwtauth.JWTAuth, email, pass string) string {
	t.Helper()

	claims, err := user.Authenticate(
		context.Background(), DB, time.Now(),
		email, pass,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, tknString, err := tokenAuth.Encode(claims)
	if err != nil {
		t.Fatal(err)
	}

	return tknString
}

// Teardown releases any resources used for the test.
func (test *Test) Teardown() {
	test.cleanup()
}

// Token generates an authenticated token for a user.
func (test *Test) Token(email, pass string) string {
	test.t.Helper()

	claims, err := user.Authenticate(
		context.Background(), test.DB, time.Now(),
		email, pass,
	)
	if err != nil {
		test.t.Fatal(err)
	}

	_, tknString, err := test.tokenAuth.Encode(claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return tknString
}
