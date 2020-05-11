package userapi

import (
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
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
	e          *httpexpect.Expect
)

var userTest *tests.Test

func TestSuite(t *testing.T) {
	userTest = tests.NewTest(t)
	t.Cleanup(userTest.Cleanup)

	r := chi.NewRouter()

	userServer := NewServer("testing", nil, userTest.Dbx)
	r.Route("/api/v1/", func(r chi.Router) {
		r.Mount("/users", userServer.Router)
	})

	userTest.SetupTestUsers(t)

	testServer := httptest.NewServer(r)
	e = httpexpect.New(t, testServer.URL)

	t.Run("get token", TestToken)
	t.Run("users get by admin", TestUsersGetByAdmin)
	t.Run("users get by user", TestUsersGetByUser)
	t.Run("users get", TestUsersGetByAdmin)
	t.Run("users", TestUsers)
}

func TestUsersGetByUser(t *testing.T) {
	errObj := e.GET("/api/v1/users").
		WithHeader("Authorization", "Bearer "+userTest.User.Token).
		Expect().
		Status(http.StatusForbidden).
		JSON().Object()

	errObj.Path("$.error.message").Equal("data is available only for admin")
}

func TestUsersGetByAdmin(t *testing.T) {

	unauthorizedErr := e.GET("/api/v1/users/").
		Expect().
		Status(http.StatusUnauthorized).
		JSON().Object()

	unauthorizedErr.Path("$.error.message").Equal("no token found")

	authAdmin := e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+userTest.Admin.Token)
	})

	userSlice := e.GET("/api/v1/users/").
		WithHeader("Authorization", "Bearer "+userTest.Admin.Token).
		Expect().
		Status(http.StatusOK).
		JSON().Array()

	userSlice.Length().Equal(4)

	// db should return 4 users
	authAdmin.GET("/api/v1/users").
		Expect().
		Status(http.StatusOK).
		JSON().Array().Length().Equal(4)

	userAdmin := authAdmin.GET("/api/v1/users/{userId}", userTest.Admin.UserID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	userAdmin.NotEmpty()

	userUser := authAdmin.GET("/api/v1/users/{userId}", userTest.User.UserID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	userUser.NotEmpty()
}

func TestUsersGet(t *testing.T) {
	tokenObject := e.GET("/api/v1/users/").
		WithBasicAuth("user@example.com", "gophers").
		Expect().
		Status(http.StatusOK).JSON().Object()

	token := tokenObject.Value("token").String().Raw()

	auth := e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+token)
	})

	// get users
	auth.GET("/api/v1/users").
		Expect().
		Status(http.StatusOK)

	e.GET("/api/v1/users/{userId}", userTest.Admin.Token).
		Expect().
		Status(http.StatusUnauthorized)
}

func TestToken(t *testing.T) {
	tokenObject := e.GET("/api/v1/users/token").
		WithBasicAuth("admin@example.com", "gophers").
		Expect().
		Status(http.StatusOK).JSON().Object()

	adminToken := tokenObject.Value("token").String().Raw()

	auth := e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+adminToken)
	})

	// get users
	auth.GET("/api/v1/users").
		Expect().
		Status(http.StatusOK)

	e.GET("/api/v1/users/{userId}", userTest.Admin.Token).
		Expect().
		Status(http.StatusUnauthorized)
}

func getTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	if userServer == nil {
		test := tests.NewIntegration(t)

		shutdown := make(chan os.Signal, 1)
		api := NewServer("test", shutdown, test.Dbx)
		userServer = httptest.NewServer(api.Router)

		//adminToken = test.Token("admin@example.com", "gophers")
		//userToken = test.Token("user@example.com", "gophers")
	}
	return userServer
}

func TestUsers(t *testing.T) {

	errObj := e.GET("/api/v1/users").
		Expect().
		Status(http.StatusUnauthorized).
		JSON().Object()

	errObj.Path("$.error.message").Equal("no token found")

	authAdmin := e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+userTest.Admin.Token)
	})

	count := authAdmin.GET("/api/v1/users").
		Expect().
		Status(http.StatusOK).
		JSON().Array().NotEmpty().Length()

	count.Equal(4)

	newUser := user.NewUser{
		Name:            "Bill Kennedy",
		Email:           "bill@ardanlabs.com",
		Roles:           []string{auth.RoleAdmin},
		Password:        "gophers",
		PasswordConfirm: "gophers",
	}

	// /api/v1/users
	newUserObj := authAdmin.POST("/api/v1/users/").
		WithJSON(newUser).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	newUserObj.ValueEqual("name", "Bill Kennedy")
	newUserObj.ValueEqual("email", "bill@ardanlabs.com")
	newUserObj.ValueEqual("roles", []string{auth.RoleAdmin})
	newUserID := newUserObj.Value("id").String().Raw()

	// /api/v1/users
	authAdmin.GET("/api/v1/users/").
		Expect().
		Status(http.StatusOK).
		JSON().Array().Length().Equal(count.Raw() + 1)

	// /api/v1/users/{userID}
	authAdmin.DELETE("/api/v1/users/{userID}", newUserID).
		Expect().
		Status(http.StatusOK)

	// /api/v1/users
	authAdmin.GET("/api/v1/users").
		Expect().
		JSON().Array().Length().Equal(count.Raw())
}
