package restaurantapi

import (
	"github.com/gavv/httpexpect/v2"
	"net/http"
	"testing"
)

const (
	menuLokys1ID = "4058d981-0df1-45de-807e-b8e90bcb2d80"
	menuLokys2ID = "f70a7f9a-e41a-47e5-b56c-444646df77bc"
)

func TestVoteAnonymous(t *testing.T) {
	server := getTestServer(t)
	e := httpexpect.New(t, server.URL)

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	respObj := e.POST("/{restaurantId}/menu/{menuId}/vote",
		restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-01").
		Expect().
		Status(http.StatusUnauthorized).
		JSON().Object()

	respObj.Value("error").Object().ValueEqual("message", "no token found")
}
// Voting for menu without token should return HTTP Status
// Unauthorised 401
func TestDoVoteWithoutToken(t *testing.T) {
	server := getTestServer(t)
	e := httpexpect.New(t, server.URL)

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	e.POST("/{restaurantId}/menu/{menuId}/vote",
		restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusUnauthorized).
		JSON().Object().
		Path("$.error.message").Equal("no token found")

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	e.POST("/{restaurantId}/menu/{menuId}/vote",
		restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-02").
		Expect().Status(http.StatusUnauthorized).
		JSON().Object().
		Path("$.error.message").Equal("no token found")
}

// GIVEN: Authenticated User Is allowed to vote once per day.
// WHEN:  The Same User votes second time for the same day
// THEN:  Second vote should be rejected
func TestVoteAuthorizedSecondPerDayForbidden(t *testing.T) {
	server := getTestServer(t)
	e := httpexpect.New(t, server.URL).
		Builder(func (req *httpexpect.Request) {
			req.WithHeader("Authorization", "Bearer "+user1Token)
		})

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	e.POST("/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusCreated).
		JSON().Object().
		ValueEqual("success", "vote accepted")

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	e.POST("/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusForbidden).
		JSON().Object().
		Path("$.error.message").
		Equal("user has already voted today")
}

func TestVoteAuthorizedTwoPerDay(t *testing.T) {
	server := getTestServer(t)
	e := httpexpect.New(t, server.URL)

	e.GET("/votes").
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusOK).JSON().Array().
		Length().Equal(1)

	e.GET("/votes").
		WithQuery("date", "2020-03-02").
		Expect().Status(http.StatusOK).JSON().Array().
		Length().Equal(1)

	e.GET("/votes").
		WithQuery("date", "2020-03-03").
		Expect().Status(http.StatusOK).JSON().Array().
		Length().Equal(0)

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	o := e.POST("/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-01").
		WithHeader("Authorization", "Bearer "+user2Token).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	o.ValueEqual("success", "vote accepted")


	e.GET("/votes").
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusOK).JSON().Array().
		Length().Gt(1)

		//Path("$.error.message").Equal("no token found")


	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	e.POST("/{restaurantId}/menu/{menuId}/vote",
		restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-01").
		WithHeader("Authorization", "Bearer "+user1Token).
		Expect().Status(http.StatusForbidden).
		JSON().Object().
		Path("$.error.message").Equal("no token found")

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	e.POST("/{restaurantId}/menu/{menuId}/vote",
		restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-02").
		WithHeader("Authorization", "Bearer "+user2Token).
		Expect().Status(http.StatusUnauthorized).
		JSON().Object().
		Path("$.error.message").Equal("no token found")

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	e.POST("/{restaurantId}/menu/{menuId}/vote",
		restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-02").
		WithHeader("Authorization", "Bearer "+user2Token).
		Expect().Status(http.StatusForbidden).
		JSON().Object().
		Path("$.error.message").Equal("no token found")
}

func TestVoteForMenu(t *testing.T) {
	server := getTestServer(t)
	e := httpexpect.New(t, server.URL)

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	voteResp := e.POST("/{restaurantId}/menu/{menuId}/vote").
		WithHeader("Authorization", "Bearer "+userToken).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	voteResp.NotEmpty()
}

func TestGetTodayVotes(t *testing.T) {
	server := getTestServer(t)
	e := httpexpect.New(t, server.URL)

	// /api/v1/restaurant/votes
	rObj := e.GET("/votes").
		Expect().Status(http.StatusOK).
		JSON().Array()

	rObj.NotEmpty()

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	rObj1 := e.GET("/votes").
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusOK).
		JSON().Array()

	rObj1.Length().Equal(1)
	rObj1.Element(0).Object().ValueEqual("date", NewDate(2020,3,1))
	rObj1.Element(0).Object().ValueEqual("id", menuLokys1ID)

	rObj2 := e.GET("/votes").
		WithQuery("date", "2020-03-02").
		Expect().Status(http.StatusOK).
		JSON().Array()

	rObj2.Length().Equal(1)
	rObj2.Element(0).Object().ValueEqual("date", NewDate(2020,3,2))
	rObj2.Element(0).Object().ValueEqual("id", menuLokys2ID)
}
