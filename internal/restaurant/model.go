package restaurant

import (
	"time"
)

// Restaurant entity stored in DB
type Restaurant struct {
	ID          string    `db:"restaurant_id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Address     string    `db:"address" json:"address"`
	OwnerUserID string    `db:"owner_user_id" json:"ownerUserId"`
	DateCreated time.Time `db:"date_created" json:"dateCreated"`
	DateUpdated time.Time `db:"date_updated" json:"dateUpdated"`
}

// NewRestaurant is what we require from clients when adding a Restaurant.
type NewRestaurant struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
	//OwnerUserID string `json:"owner_user_id" validate:"required"`
}

// UpdateRestaurant defines what information may be provided to modify an
// existing Restaurant. All fiends are optional so clients can send just the
// fields they want changed. It uses pointer fields so we can differentiate
// between a field that was not provided and field that was provided as
// explicitly blank. Normally we do not want to use pointers to basic types but
// we make exceptions around marshalling/unmarshalling.
type UpdateRestaurant struct {
	Name    *string `json:"name"`
	Address *string `json:"address"`
}

// Menu defines and entity stored in DB.
type Menu struct {
	ID           string    `db:"menu_id" json:"id"`
	RestaurantID string    `db:"restaurant_id" json:"restaurantId"`
	Date         time.Time `db:"date" json:"date"`
	Menu         string    `db:"menu" json:"menu"`
	Votes        int       `db:"votes" json:"votes"`
}

// UpdateMenu used as an incoming http data to perform menu update or menu create
type UpdateMenu struct {
	ID           string    `db:"menu_id" json:"id"`
	RestaurantID string    `db:"restaurant_id" json:"restaurantId"`
	Menu         string    `db:"menu" json:"menu"`
	Date         time.Time `db:"date" json:"date"`
}
