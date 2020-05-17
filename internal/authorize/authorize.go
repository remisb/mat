package authorize

import "github.com/remisb/mat/internal/auth"

const (
	create = "create"
	update = "update"
	delete = "delete"
)

type action string
type object int

const (
	restaurant object = iota
	menu
	vote
)

type effect int

const (
	allow effect = iota
	deny
)

func (e effect) String() string {
	switch e {
	case allow:
		return "allow"
	case deny:
		return "deny"
	}
	return ""
}

// Authorizer is the interface that performs user's authorization.
type Authorizer interface {
	Allowed(userID, role, action string)
}

type rule struct {
	effect effect
	role   string
	action string
}

// AuthorizerNaive is minimalistic and naive implementation to perform role based user's authorization.
type AuthorizerNaive struct {
	rules map[object]map[action]rule
}

// New is a factory function which create new minimalistic, naive Authorization service.
func New() *AuthorizerNaive {
	a := AuthorizerNaive{}
	a.addRule(allow, auth.RoleAdmin, restaurant, create)
	a.addRule(allow, auth.RoleAdmin, restaurant, update)
	a.addRule(allow, auth.RoleAdmin, restaurant, delete)
	a.addRule(allow, auth.RoleAdmin, menu, create)
	a.addRule(allow, auth.RoleAdmin, menu, update)
	a.addRule(allow, auth.RoleAdmin, menu, delete)
	a.addRule(allow, auth.RoleAdmin, vote, delete)
	return &a
}

func (a *AuthorizerNaive) addRule(effect effect, role string, object object, action string) {
	//a.rules[object]
	//rule{effect, role, action}
}

// Allowed checks does the user has requested role.
func (a AuthorizerNaive) Allowed(userID, role, action string) bool {
	if role == auth.RoleAdmin {
		return true
	}

	return false
}
