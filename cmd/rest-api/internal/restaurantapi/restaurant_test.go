package restaurantapi

import (
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi"
	"github.com/remisb/mat/cmd/rest-api/internal/userapi"
	"github.com/remisb/mat/internal/restaurant"
	"github.com/remisb/mat/internal/tests"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const (
	restaurantLokysID   = "5828612a-1f8a-403c-b6d1-6cb66fbf0c66"
	restaurantPaikisID  = "0ce90028-69cb-4e9c-9af0-7bbada50d5b6"
	restaurantInvalidID = "Qce90028-69cb-4e9c-9af0-7bbada50d5b6"
	restaurantNoFountID = "5cf37266-3473-4006-984f-9325122678b7"
)

var (
	httpHandler      http.Handler
	e                *httpexpect.Expect
	restaurantServer *httptest.Server
)

func TestMain(m *testing.M) {
	testSetup()
	m.Run()
	testShutdown()
}

func testSetup() {

}

func testShutdown() {

}

var restaurantTest *tests.Test

func TestSuite(t *testing.T) {
	restaurantTest = tests.NewTest(t)
	t.Cleanup(restaurantTest.Cleanup)

	r := chi.NewRouter()

	userServer := userapi.NewServer("testing", nil, restaurantTest.Dbx)
	restaurantServer := NewServer("development", nil, restaurantTest.Dbx)
	r.Route("/api/v1/", func(r chi.Router) {
		r.Mount("/users", userServer.Router)
		r.Mount("/restaurant", restaurantServer.Router)
	})

	restaurantTest.SetupTestUsers(t)

	testServer := httptest.NewServer(r)
	e = httpexpect.New(t, testServer.URL)

	t.Run("restaurants get", TestGetRestaurants)
	t.Run("restaurant get", TestGetRestaurant)
	t.Run("restaurant create", TestCreateRestaurant)
	t.Run("menus get", TestGetRestaurantMenus)
	t.Run("menu get", TestRestaurantMenuRetrieval)
	t.Run("menu create", TestCreateMenu)
	t.Run("menu update", TestUpdateMenu)

	t.Run("vote by anonymous user", TestVoteAnonymous)
	t.Run("vote get today votes by anonymous", TestGetTodayVotes)
	t.Run("vote by user 1", TestVoteTodayUser1)
	t.Run("vote second per day is forbidden", TestVoteAuthorizedSecondPerDayForbidden)
	t.Run("vote by user", TestVoteAuthorizedTwoPerDay)

}

func TestGetRestaurants(t *testing.T) {
	e.GET("/api/v1/restaurant").Expect().
		Status(http.StatusOK).
		JSON().Array().Length().Equal(5)
}

func TestGetRestaurant(t *testing.T) {
	e.GET("/api/v1/restaurant/{restaurantId}", restaurantInvalidID).
		Expect().Status(http.StatusBadRequest).
		JSON().Object().
		Path("$.error.message").Equal("ID is not in its proper form")

	restaurant1 := e.GET("/api/v1/restaurant/{restaurantID}", restaurantPaikisID).
		Expect().Status(http.StatusOK).
		JSON().Object()

	date1 := NewDate(2019, 3, 24)

	restaurant1.ValueEqual("id", restaurantPaikisID)
	restaurant1.ValueEqual("name", "Paikis")
	restaurant1.ValueEqual("address", "A. Smetonos g. 5, Vilnius 01115")
	restaurant1.ValueEqual("ownerUserId", "5cf37266-3473-4006-984f-9325122678b7")
	restaurant1.ValueEqual("dateCreated", date1)
	restaurant1.ValueEqual("dateUpdated", date1)

	restaurant2 := e.GET("/api/v1/restaurant/{restaurantID}", restaurantLokysID).
		Expect().Status(http.StatusOK).
		JSON().Object()

	restaurant2.ValueEqual("id", restaurantLokysID)
	restaurant2.ValueEqual("name", "Lokys")
	restaurant2.ValueEqual("address", "Stiklių g. 10, Vilnius 01131")
	restaurant2.ValueEqual("ownerUserId", "5cf37266-3473-4006-984f-9325122678b7")
	restaurant2.ValueEqual("dateCreated", date1)
	restaurant2.ValueEqual("dateUpdated", date1)

	e.GET("/api/v1/restaurant/{restaurantID}", restaurantNoFountID).
		Expect().Status(http.StatusNotFound).
		JSON().Object().Path("$.error.message").Equal("Restaurant not found")
}

func TestGetRestaurantMenus(t *testing.T) {

	// /api/v1/restaurant/:restaurantId/menu
	// /api/v1/restaurant/5828612a-1f8a-403c-b6d1-6cb66fbf0c66/menu
	e.GET("/api/v1/restaurant/{restaurantId}/menu", restaurantLokysID).
		Expect().
		Status(http.StatusOK).JSON().Array().
		Length().Equal(2)

	e.GET("/api/v1/restaurant/{restaurantId}/menu", restaurantPaikisID).
		Expect().
		Status(http.StatusOK).JSON().Array().
		Length().Equal(0)

	errObject := e.GET("/api/v1/restaurant/{restaurantId}/menu", restaurantInvalidID).
		Expect().
		Status(http.StatusBadRequest).JSON().Object()

	errObject.Value("error").Object().
		ValueEqual("message", "ID is not in its proper form")

	errObject = e.GET("/api/v1/restaurant/{restaurantId}/menu", restaurantNoFountID).
		Expect().
		Status(http.StatusNotFound).JSON().Object()

	errObject.Value("error").Object().
		ValueEqual("message", "Restaurant not found")
}

func getTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	if restaurantServer == nil {
		test := tests.NewIntegration(t)
		defer test.Teardown()

		shutdown := make(chan os.Signal, 1)

		api := NewServer("test", shutdown, test.Dbx)
		restaurantServer = httptest.NewServer(api.Router)
	}
	return restaurantServer
}

func TestCreateRestaurant(t *testing.T) {

	// /api/v1/restaurant
	newRestaurant := restaurant.NewRestaurant{
		Name:    "Restaurant test name 1",
		Address: "restaurant test address 1",
	}

	newRestaurant2 := restaurant.NewRestaurant{
		Name:    "Restaurant test name 2",
		Address: "restaurant test address 2",
	}

	// without claims should fail
	e.POST("/api/v1/restaurant").WithJSON(newRestaurant).
		Expect().
		Status(http.StatusUnauthorized).
		JSON().Object().
		Path("$.error.message").Equal("no token found")

	// get claims for test user 1s
	restaurantObj := e.POST("/api/v1/restaurant").
		WithHeader("Authorization", "Bearer "+restaurantTest.User.Token).
		WithJSON(newRestaurant).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	assertRestaurantError(restaurantObj, newRestaurant)

	// create restaurant with admin token
	restaurantObj = e.POST("/api/v1/restaurant").
		WithHeader("Authorization", "Bearer "+restaurantTest.Admin.Token).
		WithJSON(newRestaurant).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	assertRestaurantError(restaurantObj, newRestaurant)

	// create restaurant with user token
	restaurantObj = e.POST("/api/v1/restaurant").
		WithHeader("Authorization", "Bearer "+restaurantTest.User.Token).
		WithJSON(newRestaurant2).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	assertRestaurantError(restaurantObj, newRestaurant2)

	// with claim should succeed

	// After restaurant was created
	// test returned ownerUserID
	// test Name
	// test Address
}

func TestCreateMenu(t *testing.T) {

	newMenu := restaurant.NewMenu{
		RestaurantID: restaurantPaikisID,
		Menu:         "Menu test content 1 for 2030.03.24 for Lokys restaurant",
		Date:         NewDate(2020, 3, 24),
	}

	// admin success
	menuObj := e.POST("/api/v1/restaurant/{restaurantId}/menu", restaurantLokysID).
		WithHeader("Authorization", "Bearer "+restaurantTest.Admin.Token).
		WithJSON(newMenu).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	assertMenuEqual(menuObj, newMenu)

	newMenu2 := restaurant.NewMenu{
		RestaurantID: restaurantPaikisID,
		Menu:         "Menu test content 1 for 2030.03.24 for Lokys restaurant",
		Date:         NewDate(2020, 3, 25),
	}

	// user success
	menuObj = e.POST("/api/v1/restaurant/{restaurantId}/menu", restaurantLokysID).
		WithHeader("Authorization", "Bearer "+restaurantTest.User.Token).
		WithJSON(newMenu2).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	assertMenuEqual(menuObj, newMenu2)
}

func assertMenuEqual(actual *httpexpect.Object, expected restaurant.NewMenu) {
	actual.Value("id").NotNull()
	actual.ValueEqual("restaurantId", expected.RestaurantID)
	actual.ValueEqual("menu", expected.Menu)
	actual.ValueEqual("date", expected.Date)
}

func TestUpdateMenu(t *testing.T) {

	newMenuUpdate := restaurant.NewMenu{
		RestaurantID: restaurantLokysID,
		Menu:         "Lokys menu for 2020-03-02 updated",
		Date:         NewDate(2020, 3, 2),
	}

	// admin success
	menuObj := e.POST("/api/v1/restaurant/{restaurantId}/menu", restaurantLokysID).
		WithHeader("Authorization", "Bearer "+restaurantTest.Admin.Token).
		WithJSON(newMenuUpdate).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	// date should not be updated
	// votes count should not be updated

	menuObj.Value("id").NotNull()
	menuObj.ValueEqual("restaurantId", restaurantLokysID)
	menuObj.ValueEqual("menu", newMenuUpdate.Menu)
	menuObj.ValueEqual("date", newMenuUpdate.Date)
	menuObj.ValueEqual("votes", 0)
}

func TestRestaurantMenuRetrieval(t *testing.T) {

	e.GET("/api/v1/restaurant//menu").
		Expect().Status(http.StatusBadRequest).
		JSON().Object().
		Path("$.error.message").Equal("restaurantID is undefined")

	e.GET("/api/v1/restaurant/{restaurantId}/menu", restaurantInvalidID).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().
		Path("$.error.message").Equal("ID is not in its proper form")

	e.GET("/api/v1/restaurant/{restaurantId}/menu", restaurantLokysID).
		Expect().Status(http.StatusOK).
		JSON().Array().Length().Equal(2)

	e.GET("/api/v1/restaurant/{restaurantId}/menu", restaurantPaikisID).
		Expect().Status(http.StatusOK).
		JSON().Array().Length().Equal(0)
}

func assertRestaurantError(actual *httpexpect.Object, expected restaurant.NewRestaurant) {
	actual.ValueEqual("name", expected.Name)
	actual.ValueEqual("address", expected.Address)
	actual.Value("ownerUserId").NotNull()
	actual.Value("dateCreated").NotNull()
	actual.Value("dateUpdated").NotNull()
}

func NewDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
