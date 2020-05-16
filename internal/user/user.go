package user

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/remisb/mat/internal/db"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const (
	// RoleAdmin is used to mark user to have Admin role.
	RoleAdmin = "ADMIN"
	// RoleUser is used to mark user to have a regular User role.
	RoleUser = "USER"

	pageSize   = 10
	queryPaged = `SELECT * FROM users OFFSET $1 LIMIT $2`
	queryAll   = `SELECT * FROM users`
)

// Repo is a user Repository structure.
type Repo struct {
	db *sqlx.DB
}

// NewRepo is a factory function used to create new user Repository.
func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db}
}

// ListPaged retrieves a list of existing users from the database with pagination.
func ListPaged(ctx context.Context, db *sqlx.DB, page int) ([]User, error) {
	offset := pageSize * (page - 1)
	users := []User{}
	if err := db.SelectContext(ctx, &users, queryPaged, offset, pageSize); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}
	return users, nil
}

// GetUsers retrieves a list of existing users from the database.
func (r *Repo) GetUsers(ctx context.Context) ([]User, error) {
	users := make([]User, 0)
	if err := r.db.SelectContext(ctx, &users, queryAll); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}
	return users, nil
}

// Retrieve gets the specified user from the database.
func (r *Repo) Retrieve(ctx context.Context, id string, roles []string) (*User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, db.ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !containsRole(roles, RoleAdmin) {
		return nil, db.ErrForbidden
	}

	var u User
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := r.db.GetContext(ctx, &u, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, db.ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting user %q", id)
	}

	return &u, nil
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// Create inserts a new user into the database.
//func Create(ctx context.Context, db *sqlx.DB, n NewUser, now time.Time) (*User, error) {
func (r *Repo) Create(ctx context.Context, name, email, password string, roles []string, now time.Time) (*User, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, "generating password hash")
	}

	u := User{
		ID:           uuid.New().String(),
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Roles:        roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(user_id, name, email, password_hash, roles, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	res, err := r.db.ExecContext(
		ctx, q,
		u.ID, u.Name, u.Email,
		u.PasswordHash, u.Roles,
		u.DateCreated, u.DateUpdated,
	)
	if err != nil {
		return nil, errors.Wrap(err, "inserting user")
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "rows affected")
	}
	if rows == 0 {
		return nil, errors.New("no new rows was added to users table")
	}

	return &u, nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims value representing this user. The claims can be
// used to generate a token for future authentication.
func (r *Repo) Authenticate(ctx context.Context, email, password string) (User, error) {
	const q = `SELECT * FROM users WHERE email = $1`

	var u User
	if err := r.db.GetContext(ctx, &u, q, email); err != nil {

		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if err == sql.ErrNoRows {
			return User{}, db.ErrAuthenticationFailure
		}

		return User{}, errors.Wrap(err, "selecting single user")
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return User{}, db.ErrAuthenticationFailure
	}
	return u, nil
}

// Update replaces a user document in the database.
func (r *Repo) Update(ctx context.Context, id string, email *string, roles []string,
	password *string, now time.Time) error {
	u, err := r.Retrieve(ctx, id, roles)
	if err != nil {
		return err
	}

	if email != nil {
		u.Email = *email
	}
	if roles != nil {
		u.Roles = roles
	}
	if password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		u.PasswordHash = pw
	}

	u.DateUpdated = now

	const q = `UPDATE users SET
		"name" = $2,
		"email" = $3,
		"roles" = $4,
		"password_hash" = $5,
		"date_updated" = $6
		WHERE user_id = $1`
	_, err = r.db.ExecContext(ctx, q, id,
		u.Name, u.Email, u.Roles,
		u.PasswordHash, u.DateUpdated,
	)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete removes a user from the database.
func (r *Repo) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return db.ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id = $1`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting user %s", id)
	}

	return nil
}
