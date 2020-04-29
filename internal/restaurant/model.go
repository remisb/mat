package restaurant

import "time"

// Restaurant entity stored in DB
type Restaurant struct {
	ID          string    `db:"restaurant_id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Address     string    `db:"address" json:"address"`
	OwnerUserID string    `db:"owner_user_id" json:"owner_user_id"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
	DateUpdated time.Time `db:"date_updated" json:"date_updated"`
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

type Menu struct {
	ID           string    `db:"menu_id" json:"id"`
	RestaurantID string    `db:"restaurant_id" json:"restaurant_id"`
	Date         time.Time `db:"date" json:"date"`
	Menu         string    `db:"menu" json:"menu"`
	Votes        int       `db:"votes" json:"votes"`
}

type NewMenu struct {
	RestaurantID string    `db:"restaurant_id" json:"restaurant_id"`
	Date         time.Time `db:"date" json:"date"`
	Menu         string    `db:"menu" json:"menu"`
}

type UpdateMenu struct {
	ID   string    `db:"menu_id" json:"id"`
	Menu string    `db:"menu" json:"menu"`
	Date time.Time `db:"date" json:"date"`
}


