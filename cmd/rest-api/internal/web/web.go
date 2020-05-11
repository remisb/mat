package web

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/cors"
	"net/http"
)

const (
	HeaderContentType   = "Content-Type"
	MimeApplicationJSON = "application/json"
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

var CorsHandler = cors.Handler(cors.Options{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
	ExposedHeaders:   []string{"Link"},
	AllowCredentials: false,
	MaxAge:           300, // Maximum value not ignored by any of major browsers
})

func RespondError(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	Respond(w, r, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...)},
	})
}
func Respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set(HeaderContentType, MimeApplicationJSON)
	w.WriteHeader(status)
	if data != nil {
		EncodeBody(w, r, data)
	}
}

func EncodeBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func DecodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
