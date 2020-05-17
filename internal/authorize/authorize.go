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
	Restaurant object = iota
	Menu
	Vote
)

type Effect int

const (
	Allow Effect = iota
	Deny
)

func (e Effect) String() string {
	switch e {
	case Allow:
		return "Allow"
	case Deny:
		return "Deny"
	}
	return ""
}

type Authorizer interface {
	Allowed(userID, role, action string)
}

type rule struct {
	effect Effect
	role   string
	action string
}

type AuthorizerNaive struct {
	rules map[object]map[action]rule
}

func New() *AuthorizerNaive {
	a := AuthorizerNaive{}
	a.addRule(Allow, auth.RoleAdmin, Restaurant, create)
	a.addRule(Allow, auth.RoleAdmin, Restaurant, update)
	a.addRule(Allow, auth.RoleAdmin, Restaurant, delete)
	a.addRule(Allow, auth.RoleAdmin, Menu, create)
	a.addRule(Allow, auth.RoleAdmin, Menu, update)
	a.addRule(Allow, auth.RoleAdmin, Menu, delete)
	a.addRule(Allow, auth.RoleAdmin, Vote, delete)
	return &a
}

func (a *AuthorizerNaive) addRule(effect Effect, role string, object object, action string) {
	//a.rules[object]
	//rule{effect, role, action}
}

func (a AuthorizerNaive) Allowed(userID, role, action string) bool {
	if role == auth.RoleAdmin {
		return true
	}

	return false
}
