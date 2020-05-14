package web

import (
	"errors"
	"github.com/go-chi/jwtauth"
	"net/http"
)

var (
	ErrNoTokenFound = errors.New("no token found")
)

var Auth *jwtauth.JWTAuth

func InitAuth() {
	Auth = jwtauth.New("HS256", []byte("secret"), nil)
}

// Verifier http middleware handler will verify a JWT string from a http request.
//
// Verifier will search for a JWT token in a http request, in the order:
//   1. 'jwt' URI query parameter -- removed
//   2. 'Authorization: BEARER T' request header
//   3. Cookie 'jwt' value -- removed
//
// The first JWT string that is found as a query parameter, authorization header
// or cookie header is then decoded by the `jwt-go` library and a *jwt.Token
// object is set on the request context. In the case of a signature decoding error
// the Verifier will also set the error on the request context.
//
// The Verifier always calls the next http handler in sequence, which can either
// be the generic `jwtauth.Authenticator` middleware or your own custom handler
// which checks the request context jwt token and error to prepare a custom
// http response.
func Verifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verify(ja, jwtauth.TokenFromHeader)(next)
	}
}

// Authenticator is a default authentication middleware to enforce access from the
// Verifier middleware request context values. The Authenticator sends a 401 Unauthorized
// response for any unverified tokens and passes the good ones through. It's just fine
// until you decide to write something similar and customize your client response.
func Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			if err == jwtauth.ErrNoTokenFound {
				RespondError(w, r, http.StatusUnauthorized, ErrNoTokenFound)
				return
			}
			RespondError(w, r, http.StatusUnauthorized, err)
			return
		}

		if token == nil || !token.Valid {
			err := errors.New("no token passed or token is invalid")
			RespondError(w, r, http.StatusUnauthorized, err)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}
