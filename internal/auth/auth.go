package auth

import (
	"context"
	"github.com/go-chi/jwtauth"
	"github.com/remisb/mat/internal/user"
	"time"
)

type Authenticator interface {
	NewToken(ctx context.Context, email, password string) (string, user.User, error)
	Authenticate(ctx context.Context, email, password string) (Claims, user.User, error)
}

type DefaultAuthenticator struct {
	tokenAuth *jwtauth.JWTAuth
	userRepo  *user.Repo
}

func (a DefaultAuthenticator) NewToken(ctx context.Context, email, password string) (string, user.User, error) {

	authenticatedUser, err := a.userRepo.Authenticate(ctx, email, password)
	if err != nil {
		return "", authenticatedUser, err
	}

	claims := NewClaims(authenticatedUser.ID,
		authenticatedUser.Name, authenticatedUser.Email,
		authenticatedUser.Roles, time.Now(), time.Hour)

	_, tokenString, _ := a.tokenAuth.Encode(claims)
	return tokenString, authenticatedUser, err
}

func (a DefaultAuthenticator) Authenticate(ctx context.Context, email, password string) (Claims, user.User, error) {
	// get user struct
	authenticatedUser, err := a.userRepo.Authenticate(ctx, email, password)
	if err != nil {
		return Claims{}, authenticatedUser, err
	}

	// convert user struct into claim
	claims := NewClaims(authenticatedUser.ID, authenticatedUser.Name, authenticatedUser.Email,
		authenticatedUser.Roles, time.Now(), time.Hour)
	return claims, authenticatedUser, nil
}

func New(userRepo *user.Repo, auth *jwtauth.JWTAuth) *Authenticator {
	da := DefaultAuthenticator{
		tokenAuth: auth,
		userRepo:  userRepo,
	}
	var a Authenticator = da
	return &a
}
