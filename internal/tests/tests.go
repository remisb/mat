package tests

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/db/dbtest"
	logz "github.com/remisb/mat/internal/log"
	"github.com/remisb/mat/internal/user"
	"go.uber.org/zap"

	"github.com/remisb/mat/internal/schema"
	"testing"
	"time"
)

type Test struct {
	DB            *sqlx.DB
	Log           *zap.SugaredLogger
	Authenticator *auth.Authenticator
	t             *testing.T
	cleanup       func()
	UserToken     string
	AdminToken    string
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

func NewIntegration(t *testing.T) *Test {
	t.Helper()

	db, cleanup := NewUnit(t)
	if err := schema.Seed(db); err != nil {
		t.Fatal(err)
	}

	// Create the logger to use.
	//logger := log.New(os.Stdout, "TEST : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Create RSA keys to enable authentication in our service.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Build an authenticator using this static key.
	kid := "4754d86b-7a6d-4df5-9c65-224741361492"
	kf := auth.NewSimpleKeyLookupFunc(kid, key.Public().(*rsa.PublicKey))
	authenticator, err := auth.NewAuthenticator(key, kid, "RS256", kf)
	if err != nil {
		t.Fatal(err)
	}

	userToken := Token(t, db, authenticator, "user@example.com", "gophers")
	adminToken := Token(t, db, authenticator, "admin@example.com", "gophers")

	return &Test{
		DB:            db,
		Log:           logz.Sugar,
		Authenticator: authenticator,
		t:             t,
		cleanup:       cleanup,
		UserToken:     userToken,
		AdminToken:    adminToken,
	}
}

// Token generates an authenticated token for a user.
func Token(t *testing.T, DB *sqlx.DB, authenticator *auth.Authenticator, email, pass string) string {
	t.Helper()

	claims, err := user.Authenticate(
		context.Background(), DB, time.Now(),
		email, pass,
	)
	if err != nil {
		t.Fatal(err)
	}

	tkn, err := authenticator.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}

	return tkn
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

	tkn, err := test.Authenticator.GenerateToken(claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return tkn
}
