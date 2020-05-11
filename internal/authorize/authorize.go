package authorize

import "github.com/remisb/mat/internal/auth"

const (
	CREATE = "create"
	UPDATE = "update"
	DELETE = "delete"
)

type Action string
type Object int

const (
	Restaurant Object = iota
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
	rules map[Object]map[Action]rule
}

func New() *AuthorizerNaive {
	a := AuthorizerNaive{}
	a.addRule(Allow, auth.RoleAdmin, Restaurant, CREATE)
	a.addRule(Allow, auth.RoleAdmin, Restaurant, UPDATE)
	a.addRule(Allow, auth.RoleAdmin, Restaurant, DELETE)
	a.addRule(Allow, auth.RoleAdmin, Menu, CREATE)
	a.addRule(Allow, auth.RoleAdmin, Menu, UPDATE)
	a.addRule(Allow, auth.RoleAdmin, Menu, DELETE)
	a.addRule(Allow, auth.RoleAdmin, Vote, DELETE)
	return &a
}

func (a *AuthorizerNaive) addRule(effect Effect, role string, object Object, action string) {
	//a.rules[object]
	//rule{effect, role, action}
}

func (a AuthorizerNaive) Allowed(userID, role, action string) bool {
	if role == auth.RoleAdmin {
		return true
	}

	return false
}
