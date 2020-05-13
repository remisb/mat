package tests

import (
	"context"
	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/db/dbtest"
	logz "github.com/remisb/mat/internal/log"
	"github.com/remisb/mat/internal/restaurant"
	"github.com/remisb/mat/internal/user"
	"go.uber.org/zap"

	"github.com/remisb/mat/internal/schema"
	"testing"
	"time"
)

// Test structure is used to perform Integration tests.
type Test struct {
	Dbx            *sqlx.DB
	authenticator  auth.Authenticator
	userRepo       *user.Repo
	restaurantRepo *restaurant.Repo
	Log            *zap.SugaredLogger
	t              *testing.T
	Cleanup        func()
	tokenAuth      *jwtauth.JWTAuth
	Admin          UserToken
	User           UserToken
	User1          UserToken
	User2          UserToken
}

type UserToken struct {
	Token  string
	UserID string
}

func NewTest(t *testing.T) *Test {
	t.Helper()

	db, teardown := setupTestDbContainer(t)
	if err := schema.Seed(db); err != nil {
		t.Fatal(err)
	}

	userRepo := user.NewRepo(db)
	jwauth := jwtauth.New("HS256", []byte("secret"), nil)
	authenticator := auth.New(userRepo, jwauth)
	return &Test{
		Dbx:            db,
		userRepo:       userRepo,
		authenticator:  *authenticator,
		restaurantRepo: restaurant.NewRepo(db),
		Log:            logz.Sugar,
		t:              t,
		Cleanup:        teardown,
	}
}

func (test *Test) SetupTestUsers(t *testing.T) {
	t.Helper()

	test.Admin = test.NewToken(t, "admin@example.com", "gophers")
	test.User = test.NewToken(t, "user@example.com", "gophers")
	test.User1 = test.NewToken(t, "user1@example.com", "gophers")
	test.User2 = test.NewToken(t, "user2@example.com", "gophers")
}

// newToken generates an authenticated token for a user.
func (test *Test) NewToken(t *testing.T, email, pass string) UserToken {
	ctx := context.Background()
	token, user, err := test.authenticator.NewToken(ctx, email, pass)
	if err != nil {
		t.Fatalf("authenticate error: %v", err)
	}
	return UserToken{token, user.ID}
}

func setupTestDbContainer(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()

	c := dbtest.StartContainer(t)
	db := openDB(t, c)

	teardown := func() {
		t.Helper()
		db.Close()
		dbtest.StopContainer(t, c)
	}

	return db, teardown
}

func openDB(t *testing.T, c *dbtest.Container) *sqlx.DB {
	t.Helper()

	dbx, err := db.Open(db.Config{
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
		pingError = dbx.Ping()
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

	if err := schema.Migrate(dbx); err != nil {
		dbtest.StopContainer(t, c)
		t.Fatalf("migrating: %s", err)
	}
	return dbx
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

	return &Test{
		Log:     logz.Sugar,
		t:       t,
		Cleanup: cleanup,
	}
}

// Token generates an authenticated token for a user.
func Token(t *testing.T, DB *sqlx.DB, tokenAuth *jwtauth.JWTAuth, email, pass string) string {
	t.Helper()

	//claims, err := user.Authenticate(
	//	context.Background(), DB, time.Now(),
	//	email, pass,
	//)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//_, tknString, err := tokenAuth.Encode(claims)
	//if err != nil {
	//	t.Fatal(err)
	//}

	return ""
}

// Teardown releases any resources used for the test.
func (test *Test) Teardown() {
	test.Cleanup()
}

// Token generates an authenticated token for a user.
func (test *Test) Token(email, pass string) string {
	test.t.Helper()
	var token string

	//claims, err := user.Authenticate(
	//	context.Background(), test.DB, time.Now(),
	//	email, pass,
	//)
	//if err != nil {
	//	test.t.Fatal(err)
	//}
	//
	//_, tknString, err := test.tokenAuth.Encode(claims)
	//if err != nil {
	//	test.t.Fatal(err)
	//}
	//
	//return tknString

	return token
}
