package restaurant

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/remisb/mat/internal/auth"
	"time"
)

func List(ctx context.Context, db *sqlx.DB) ([]Restaurant, error) {

	restaurants := []Restaurant{}
	const q = `SELECT * FROM restaurant`
	if err := db.SelectContext(ctx, &restaurants, q); err != nil {
		return nil, errors.Wrap(err, "selecting restaurants")
	}
	return restaurants, nil
}

func Create(ctx context.Context, db *sqlx.DB, user auth.Claims, nr NewRestaurant, now time.Time) (*Restaurant, error) {
	currentTime := now.UTC()
	r := Restaurant{
		ID:          uuid.New().String(),
		Name:        nr.Name,
		Address:     nr.Address,
		OwnerUserID: user.Subject,
		DateCreated: currentTime,
		DateUpdated:  currentTime,
	}

	const q = `INSERT INTO restaurant
	    (restaurant_id, name, address, owner_user_id, date_created, date_updated)
	    VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.ExecContext(ctx, q, r.ID, r.Name, r.Address, r.OwnerUserID, r.DateCreated, r.DateUpdated)
	if err != nil {
		return nil, errors.Wrap(err, "inserting restaurant")
	}

	return &r, nil
}
