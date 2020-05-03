package test

import (
	"github.com/gavv/httpexpect/v2"
	"github.com/remisb/mat/internal/tests"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// UserTests holds methods for each user subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type UserTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

func TestUsersOld(t *testing.T) {
	test := tests.NewIntegration(t)
	defer test.Teardown()

	shutdown := make(chan os.Signal, 1)
	tests := UserTests{
		app:        Server.API("develop", shutdown, test.Log, test.DB, test.Authenticator),
		userToken:  test.Token("user@example.com", "gophers"),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	//t.Run("getToken401", tests.getToken401)
	//t.Run("getToken200", tests.getToken200)
	//t.Run("postUser400", tests.postUser400)
	//t.Run("postUser401", tests.postUser401)
	//t.Run("postUser403", tests.postUser403)
}

func TestUsers(t *testing.T) {
	test := tests.NewIntegration(t)
	defer test.Teardown()

	users := userMap{}
	mux :=
	Server{}

	handler := UsersHandler()

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.New(t, server.URL)
}

