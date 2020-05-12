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

//func TestVoteAnonymous(t *testing.T) {
//
//	e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote",
//		restaurantLokysID, menuLokys1ID).
//		WithQuery("date", "2020-03-01").
//		Expect().Status(http.StatusUnauthorized).
//		JSON().Object().
//		Path("$.error.message").Equal("no token found")
//}

// Voting for menu without token should return HTTP Status
// Unauthorised 401
func TestVoteAnonymous(t *testing.T) {

	e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusUnauthorized).
		JSON().Object().
		Path("$.error.message").Equal("no token found")

	e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-02").
		Expect().Status(http.StatusUnauthorized).
		JSON().Object().
		Path("$.error.message").Equal("no token found")
}

func TestGetTodayVotes(t *testing.T) {

	rObj := e.GET("/api/v1/restaurant/votes").
		Expect().Status(http.StatusOK).
		JSON().Array()

	rObj.Length().Equal(0)

	// /api/v1/restaurant/{restaurantId}/menu/{menuId}/vote
	rObj1 := e.GET("/api/v1/restaurant/votes").
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusOK).
		JSON().Array()

	rObj1.Length().Equal(1)
	el1 := rObj1.Element(0).Object()
	el1.ValueEqual("date", NewDate(2020, 3, 1))
	el1.ValueEqual("id", menuLokys1ID)

	rObj2 := e.GET("/api/v1/restaurant/votes").
		WithQuery("date", "2020-03-02").
		Expect().Status(http.StatusOK).
		JSON().Array()

	rObj2.Length().Equal(1)
	el2 := rObj2.Element(0).Object()
	el2.ValueEqual("date", NewDate(2020, 3, 2))
	el2.ValueEqual("id", menuLokys2ID)
}

func TestVoteTodayUser1(t *testing.T) {

	voteResp := e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithHeader("Authorization", "Bearer "+restaurantTest.User1.Token).
		Expect().Status(http.StatusCreated).
		JSON().Object()

	voteResp.NotEmpty()
}

// GIVEN: Authenticated User Is allowed to vote once per day.
// WHEN:  The Same User votes second time for the same day
// THEN:  Second vote should be rejected
func TestVoteAuthorizedSecondPerDayForbidden(t *testing.T) {

	authUser1 := e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", "Bearer "+restaurantTest.User1.Token)
	})

	authUser1.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-13").
		Expect().Status(http.StatusCreated).
		JSON().Object().
		ValueEqual("success", "vote accepted")

	authUser1.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-13").
		Expect().Status(http.StatusForbidden).
		JSON().Object().
		Path("$.error.message").Equal("user has already voted today")
}

func TestVoteAuthorizedTwoPerDay(t *testing.T) {

	e.GET("/api/v1/restaurant/votes").
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusOK).
		JSON().Array().Length().Equal(1)

	e.GET("/api/v1/restaurant/votes").
		WithQuery("date", "2020-03-02").
		Expect().Status(http.StatusOK).
		JSON().Array().Length().Equal(1)

	e.GET("/api/v1/restaurant/votes").
		WithQuery("date", "2020-03-03").
		Expect().Status(http.StatusOK).
		JSON().Array().Length().Equal(0)

	e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys1ID).
		WithQuery("date", "2020-03-01").
		WithHeader("Authorization", "Bearer "+restaurantTest.User2.Token).
		Expect().Status(http.StatusCreated).
		JSON().Object().ValueEqual("success", "vote accepted")

	e.GET("/api/v1/restaurant/votes").
		WithQuery("date", "2020-03-01").
		Expect().Status(http.StatusOK).
		JSON().Array().Length().Gt(1)

	//Path("$.error.message").Equal("no token found")

	e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-01").
		WithHeader("Authorization", "Bearer "+restaurantTest.User1.Token).
		Expect().Status(http.StatusForbidden).
		JSON().Object().
		Path("$.error.message").Equal("no token found")

	e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-02").
		WithHeader("Authorization", "Bearer "+restaurantTest.User2.Token).
		Expect().Status(http.StatusUnauthorized).
		JSON().Object().
		Path("$.error.message").Equal("no token found")

	e.POST("/api/v1/restaurant/{restaurantId}/menu/{menuId}/vote", restaurantLokysID, menuLokys2ID).
		WithQuery("date", "2020-03-02").
		WithHeader("Authorization", "Bearer "+restaurantTest.User2.Token).
		Expect().Status(http.StatusForbidden).
		JSON().Object().
		Path("$.error.message").Equal("no token found")
}
