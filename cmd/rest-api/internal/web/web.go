package web

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/cors"
	"net/http"
)

const (
	headerContentType   = "Content-Type"
	mimeApplicationJSON = "application/json"
)

// Error is used to pass an error during the request through the
// application with web specific context.
type Error struct {
	Err    error
	Status int
	Fields []FieldError
}

// FieldError is used to indicate an error with a specific request field.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &Error{err, status, nil}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (err *Error) Error() string {
	return err.Err.Error()
}

// CorsHandler has default cors settings for HTTP Middleware.
var CorsHandler = cors.Handler(cors.Options{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	ExposedHeaders:   []string{"Link"},
	AllowCredentials: false,
	MaxAge:           300, // Maximum value not ignored by any of major browsers
})

// RespondError create json error response and outputs passed error into response body.
func RespondError(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	Respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...)},
	})
}

// Respond create json response and outputs json representation of the passed data into response body.
func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set(headerContentType, mimeApplicationJSON)
	w.WriteHeader(status)
	if data != nil {
		EncodeBody(w, r, data)
	}
}

// EncodeBody encodes passed date to json format and writes it into Response body.
func EncodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

// DecodeBody decode json from request body into passed pointer struct.
func DecodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
