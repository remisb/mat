package restaurant

import (
	"context"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
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

var ErrRestaurantNotFound = errors.New("Restaurant not found")

// ListPaged is used to return paged list of Restaurants.
func ListPaged(ctx context.Context, db *sqlx.DB, page int) ([]Restaurant, error) {
	offset := pageSize * (page - 1)
	restaurants := []Restaurant{}
	err := db.SelectContext(ctx, &restaurants, queryPaged, offset, pageSize)
	if err != nil {
		return nil, errors.Wrap(err, "selecting restaurants")
	}
	return restaurants, nil
}

// RetrieveRestaurantList retrieves list of restaurants from database
func RetrieveRestaurantList(ctx context.Context, dbx *sqlx.DB) ([]Restaurant, error) {

	restaurants := []Restaurant{}
	const q = `SELECT * FROM restaurant`
	if err := dbx.SelectContext(ctx, &restaurants, q); err != nil {
		return nil, errors.Wrap(err, "selecting restaurants")
	}
	return restaurants, nil
}

// RetrieveRestaurant gets the specified user from the database.
func RetrieveRestaurant(ctx context.Context, dbx *sqlx.DB, restaurantID string) (*Restaurant, error) {
	if _, err := uuid.Parse(restaurantID); err != nil {
		return nil, db.ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	//if !claims.HasRole(auth.RoleAdmin) && claims.Subject != id {
	//	return nil, db.ErrForbidden
	//}

	var r Restaurant
	const q = `SELECT * FROM restaurant WHERE restaurant_id = $1`
	if err := dbx.GetContext(ctx, &r, q, restaurantID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRestaurantNotFound
		}

		return nil, errors.Wrapf(err, "selecting restaurant %q", restaurantID)
	}

	return &r, nil
}

// CreateRestaurant inserts new restaurant into the database.
//func CreateRestaurant(ctx context.Context, claims auth.Claims, db *sqlx.DB, nr NewRestaurant, now time.Time) (*Restaurant, error) {
func CreateRestaurant(ctx context.Context, claims jwt.MapClaims, dbx *sqlx.DB, nr NewRestaurant, now time.Time) (*Restaurant, error) {
	currentTime := now.UTC()
	r := Restaurant{
		ID:          uuid.New().String(),
		Name:        nr.Name,
		Address:     nr.Address,
		OwnerUserID: claims["sub"].(string),
		DateCreated: currentTime,
		DateUpdated: currentTime,
	}

	const q = `INSERT INTO restaurant
	    (restaurant_id, name, address, owner_user_id, date_created, date_updated)
	    VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := dbx.ExecContext(ctx, q, r.ID, r.Name, r.Address, r.OwnerUserID, r.DateCreated, r.DateUpdated)
	if err != nil {
		return nil, errors.Wrap(err, "inserting restaurant")
	}

	return &r, nil
}

// DeleteRestaurant removes a restaurant from the database.
func DeleteRestaurant(ctx context.Context, dbx *sqlx.DB, restaurantID string) error {
	if _, err := uuid.Parse(restaurantID); err != nil {
		return db.ErrInvalidID
	}

	const q = `DELETE FROM restaurant WHERE restaurant_id = $1`
	if _, err := dbx.ExecContext(ctx, q, restaurantID); err != nil {
		return errors.Wrapf(err, "deleting restaurant %s", restaurantID)
	}

	return nil
}
