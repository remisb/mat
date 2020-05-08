package userapi

import (
	"github.com/gavv/httpexpect/v2"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/tests"
	"github.com/remisb/mat/internal/user"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	userServer *httptest.Server
)

func TestToken(t *testing.T) {
	server := getTestServer(t)
	defer server.Close()

	e := httpexpect.New(t, server.URL)

	tokenObject := e.GET("/api/v1/user/token2").
		WithBasicAuth("user@example.com","gophers").
		Expect().
		Status(http.StatusOK).JSON().Object()

	token := tokenObject.Value("token").String().Raw()

	auth := e.Builder(func (req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer " + token)
	})

	// get users
	auth.GET("/").
		Expect().
 		Status(http.StatusOK)

	e.GET("/userId").
		Expect().
		Status(http.StatusUnauthorized)
}

func getTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	if userServer == nil {
		test := tests.NewIntegration(t)
		shutdown := make(chan os.Signal, 1)
		api := NewServer("test", shutdown, test.DB)
		userServer = httptest.NewServer(api.Router)

		//adminToken = test.Token("admin@example.com", "gophers")
		//userToken = test.Token("user@example.com", "gophers")
	}
	return userServer
}

func TestUsers(t *testing.T) {
	test := tests.NewIntegration(t)
	shutdown := make(chan os.Signal, 1)
	userAPI := NewServer("test", shutdown, test.DB)
	server := httptest.NewServer(userAPI.Router)
	defer server.Close()

	e := httpexpect.New(t, server.URL)

	// /api/v1/users
	count := e.GET("/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().NotEmpty().Length()

	count.Equal(2)

	// add new user
	newUser := user.NewUser{
		Name:            "Bill Kennedy",
		Email:           "bill@ardanlabs.com",
		Roles:           []string{auth.RoleAdmin},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}
	// /api/v1/users
	newUserObj := e.POST("/").
		WithJSON(newUser).
		WithHeader("Authorization", "Bearer "+test.AdminToken).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	newUserObj.ContainsKey("name").ValueEqual("name", "Bill Kennedy")
	newUserObj.ContainsKey("email").ValueEqual("email", "bill@ardanlabs.com")
	newUserObj.ContainsKey("roles").ValueEqual("roles", []string{auth.RoleAdmin})
	newUserID := newUserObj.Value("id").String()

	t.Logf("obj: %+v\n", newUserID.Raw())

	// /api/v1/users
	e.GET("/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Array().Length().Equal(count.Raw() + 1)

	// /api/v1/users/{userID}
	e.DELETE("/{userID}", newUserID.Raw()).
		Expect().Status(http.StatusOK)

	// /api/v1/users
	e.GET("/").Expect().
		JSON().
		Array().Length().Equal(count.Raw())
}
