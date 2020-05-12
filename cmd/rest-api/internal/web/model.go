package web

type TokenResult struct {
	Token string `json:"token"`
}

// APIError example
type APIError = map[string]interface{}
