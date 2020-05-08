package schema

import (
	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

// Migrate is used to perform database migration.
func Migrate(db *sqlx.DB) error {
	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
	d := darwin.New(driver, migrations, nil)
	return d.Migrate()
}

var migrations = []darwin.Migration{
	{
		Version:     1,
		Description: "Add restaurant", // Add restaurant
		Script: `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE restaurant (
	restaurant_id UUID NOT NULL DEFAULT uuid_generate_v4(),
	name          TEXT NOT NULL,
	address       TEXT,
    owner_user_id TEXT NOT NULL,
	date_created  TIMESTAMP,
	date_updated  TIMESTAMP,
	PRIMARY KEY (restaurant_id)
);`},
	{
		Version:     2,
		Description: "Add menu",
		Script: `
CREATE TABLE menu (
		menu_id       UUID NOT NULL DEFAULT uuid_generate_v4(),
		restaurant_id UUID,
		date          DATE NOT NULL DEFAULT CURRENT_DATE,
		menu          VARCHAR(1024),
		votes         INTEGER,
        PRIMARY KEY(menu_id)
)`},
	{
		Version:     3,
		Description: "Add votes",
		Script: `
CREATE TABLE vote (
    date          TIMESTAMP NOT NULL,
    user_id       UUID,
	restaurant_id UUID,
	time_voted    TIMESTAMP,

	PRIMARY KEY (date, user_id),
	FOREIGN KEY (restaurant_id) REFERENCES restaurant(restaurant_id)
);`},
	{
		Version:     4,
		Description: "Add users",
		Script: `
CREATE TABLE users (
	user_id       UUID,
	name          TEXT,
	email         TEXT UNIQUE,
	roles         TEXT[],
	password_hash TEXT,
	date_created TIMESTAMP,
	date_updated TIMESTAMP,
	PRIMARY KEY (user_id)
);`},
}
