package restaurant

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/remisb/mat/internal/db"
	"time"
)

const (
	pageSize   = 10
	queryPaged = `SELECT * FROM restaurant OFFSET $1 LIMIT $2`
	queryAll   = `SELECT * FROM restaurant`
)

// ErrRestaurantNotFound returned when restaurant is not found
var ErrRestaurantNotFound = errors.New("Restaurant not found")

// Repo is a restaurant Repository structure.
type Repo struct {
	db *sqlx.DB
}

// NewRepo is a factory function used to create new restaurant Repository.
func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db}
}

// GetRestaurantsPaged retrieves a list of existing restaurants from the database with pagination.
func (r *Repo) GetRestaurantsPaged(ctx context.Context, page int) ([]Restaurant, error) {
	offset := pageSize * (page - 1)
	restaurants := make([]Restaurant, 0)
	err := r.db.SelectContext(ctx, &restaurants, queryPaged, offset, pageSize)
	if err != nil {
		return nil, errors.Wrap(err, "selecting restaurants")
	}
	return restaurants, nil
}

// GetRestaurants retrieves a list of existing restaurants from the database.
func (r *Repo) GetRestaurants(ctx context.Context) ([]Restaurant, error) {
	restaurants := make([]Restaurant, 0)
	const q = `SELECT * FROM restaurant`
	if err := r.db.SelectContext(ctx, &restaurants, q); err != nil {
		return nil, errors.Wrap(err, "selecting restaurants")
	}
	return restaurants, nil
}

// GetRestaurant gets the specified user from the database.
func (r *Repo) GetRestaurant(ctx context.Context, restaurantID string) (*Restaurant, error) {
	if _, err := uuid.Parse(restaurantID); err != nil {
		return nil, db.ErrInvalidID
	}

	var restaurant Restaurant

	const q = `SELECT * FROM restaurant WHERE restaurant_id = $1`
	if err := r.db.GetContext(ctx, &restaurant, q, restaurantID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRestaurantNotFound
		}
		return nil, errors.Wrapf(err, "selecting restaurant %q", restaurantID)
	}

	return &restaurant, nil
}

// RetrieveRestaurantList retrieves list of restaurants from database
func (r *Repo) RetrieveRestaurantList(ctx context.Context) ([]Restaurant, error) {

	restaurants := make([]Restaurant, 0)
	const q = `SELECT * FROM restaurant`
	if err := r.db.SelectContext(ctx, &restaurants, q); err != nil {
		return nil, errors.Wrap(err, "selecting restaurants")
	}
	return restaurants, nil
}

// CreateRestaurant inserts new restaurant into the database.
//func CreateRestaurant(ctx context.Context, claims auth.Claims, db *sqlx.DB, nr NewRestaurant, now time.Time) (*Restaurant, error) {
func (r *Repo) CreateRestaurant(ctx context.Context, nr NewRestaurant, now time.Time, userID string) (*Restaurant, error) {
	currentTime := now.UTC()
	rest := Restaurant{
		ID:          uuid.New().String(),
		Name:        nr.Name,
		Address:     nr.Address,
		OwnerUserID: userID,
		DateCreated: currentTime,
		DateUpdated: currentTime,
	}

	const q = `INSERT INTO restaurant
	    (restaurant_id, name, address, owner_user_id, date_created, date_updated)
	    VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, q, rest.ID, rest.Name, rest.Address, rest.OwnerUserID, rest.DateCreated, rest.DateUpdated)
	if err != nil {
		return nil, errors.Wrap(err, "inserting restaurant")
	}

	return &rest, nil
}

// DeleteRestaurant removes a restaurant from the database.
func (r *Repo) DeleteRestaurant(ctx context.Context, restaurantID string) error {
	if _, err := uuid.Parse(restaurantID); err != nil {
		return db.ErrInvalidID
	}

	const q = `DELETE FROM restaurant WHERE restaurant_id = $1`
	if _, err := r.db.ExecContext(ctx, q, restaurantID); err != nil {
		return errors.Wrapf(err, "deleting restaurant %s", restaurantID)
	}

	return nil
}
