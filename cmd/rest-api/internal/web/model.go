package web

// TokenResult structure is used to return newly generated token from HTTP REST API endpoint.
type TokenResult struct {
	Token string `json:"token"`
}

// APIError example
type APIError = map[string]interface{}
