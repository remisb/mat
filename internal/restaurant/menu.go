package restaurant

import (
	"context"
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/remisb/mat/internal/auth"
	"github.com/remisb/mat/internal/db"
	"github.com/remisb/mat/internal/log"
	"strings"
	"time"
)

var ErrNotFound = errors.New("menu not found")

// RetrieveMenu used to retrieve menu from DB by specified menuID
func (r *Repo) RetrieveMenu(ctx context.Context, menuID string) (*Menu, error) {
	if _, err := uuid.Parse(menuID); err != nil {
		return nil, db.ErrInvalidID
	}

	var m Menu
	const q = `SELECT * FROM menu WHERE menu_id = $1`
	if err := r.db.GetContext(ctx, &m, q, menuID); err != nil {
		if err == sql.ErrNoRows {
			return nil, db.ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting menu %q", menuID)
	}

	return &m, nil
}

func (r *Repo) readMenuByRestaurantDate(ctx context.Context, restaurantID string, date time.Time) (*Menu, error) {
	var m Menu
	const q = `SELECT * FROM menu 
	    WHERE restaurant_id = $1 AND date = $2`
	if err := r.db.GetContext(ctx, &m, q, restaurantID, date); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err,
			"selecting menu for restaurant_id: %s date: %v", restaurantID, date)
	}

	return &m, nil
}

// RetrieveRestaurantMenus retrieves specified menu from database
func (r *Repo) RetrieveRestaurantMenus(ctx context.Context, restaurantID, menuID string) (*Menu, error) {
	if _, err := uuid.Parse(restaurantID); err != nil {
		return nil, db.ErrInvalidID
	}

	if _, err := uuid.Parse(menuID); err != nil {
		return nil, db.ErrInvalidID
	}

	var menu Menu
	const q = `SELECT * FROM menu WHERE restaurant_id = $1 AND menu_id = $2`
	if err := r.db.GetContext(ctx, &menu, q, restaurantID, menuID); err != nil {
		if err == sql.ErrNoRows {
			return nil, db.ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting memu %q", menuID)
	}

	return &menu, nil
}

// RetrieveMenusByDate retrieves a list of menus from DB for specified date.
// NOTE functionality of this func is identical to MenuVotes.
// QUESTION 1: Do I have to have those two functions?
// QUESTION 2: Should I modify MenuVotes functionality to be more specific for votes data retrieval?
func (r *Repo) RetrieveMenusByDate(ctx context.Context, date time.Time) ([]Menu, error) {
	var menus = make([]Menu, 0)
	const q = `SELECT * FROM menu WHERE date = $1`
	if err := r.db.SelectContext(ctx, &menus, q, date); err != nil {
		return nil, errors.Wrap(err, "retrieving menus for specified date")
	}
	return menus, nil
}

// RetrieveMenusByRestaurant retrieves list of menus from DB for specified restaurant.
func (r *Repo) RetrieveMenusByRestaurant(ctx context.Context, restaurantID string) ([]Menu, error) {
	if _, err := uuid.Parse(restaurantID); err != nil {
		return nil, db.ErrInvalidID
	}

	_, err := r.GetRestaurant(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	var menus = make([]Menu, 0)
	const q = `SELECT * FROM menu WHERE restaurant_id = $1 ORDER BY date DESC `
	if err := r.db.SelectContext(ctx, &menus, q, restaurantID); err != nil {
		return nil, errors.Wrap(err, "retrieving restaurant menus")
	}
	return menus, nil
}

// MenuVotes retrieves list of menu with votes for specified date from database
func (r *Repo) MenuVotes(ctx context.Context, date time.Time) ([]Menu, error) {
	var menus = make([]Menu, 0)
	const q = `SELECT * FROM menu WHERE date = $1`
	if err := r.db.SelectContext(ctx, &menus, q, date); err != nil {
		return nil, errors.Wrap(err, "retrieving menu votes")
	}
	return menus, nil
}

func (r *Repo) MenuVote(ctx context.Context, claims jwt.MapClaims, restaurantID, menuID string, date time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	userID := claims["sub"].(string)

	err = txMenuVote(ctx, tx, restaurantID, menuID, userID, date)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			err = errors.Wrap(err, "transaction rollback error")
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Sugar.Errorf("error on tx commit", err)
	}
	return nil
}

func txMenuVote(ctx context.Context, tx *sql.Tx, restaurantID, menuID, userID string, date time.Time) error {
	var count int
	const qSelectVote = `SELECT COUNT(*) as count FROM vote WHERE time_voted = $1 AND user_id = $2`
	err := tx.QueryRow(qSelectVote, date, userID).Scan(&count)
	if err != nil {
		return errors.Wrap(err, "error on vote count scan")
	}
	if count > 0 {
		return errors.New("user has already voted today")
	}

	const qInsertVote = `INSERT INTO vote (date, user_id, restaurant_id, time_voted)
	    VALUES ($1, $2, $3, $4)`
	voteResult, err := tx.ExecContext(ctx, qInsertVote, date, userID, restaurantID, date)
	if err != nil {
		return errors.Wrap(err, "inserting restaurant")
	}
	if count, err := voteResult.RowsAffected(); err != nil || count < 1 {
		return errors.Wrap(err, "error on getting rows updated")
	}

	const qUpdateMenuVote = `UPDATE menu SET votes = votes + 1 WHERE menu_id = $1`
	updateResult, err := tx.Exec(qUpdateMenuVote, menuID)
	if err != nil {
		return errors.Wrapf(err, "error on menu vote update")
	}

	if count, err := updateResult.RowsAffected(); err != nil || count == 0 {
		return errors.Wrapf(err, "error on getting rows updated")
	}
	return nil
}

//
func isRestaurantOwner(restaurant *Restaurant, claims jwt.MapClaims) bool {
	if restaurant == nil {
		return false
	}

	return restaurant.OwnerUserID == claims["sub"].(string)
}

func (r *Repo) MenuUpdate(ctx context.Context, claims jwt.MapClaims, dbx *sqlx.DB, um UpdateMenu) (*Menu, error) {

	restaurant, err := r.GetRestaurant(ctx, um.RestaurantID)
	if err != nil {
		return nil, errors.Wrap(db.ErrNotFound, "unable to find restaurant with restaurantID: "+um.RestaurantID)
	}

	if restaurant == nil {
		return nil, err
	}

	// does the menu exist for specified date
	menu, err := r.RetrieveMenu(ctx, um.ID)
	if err != nil {
		return nil, err
	}

	// user permission check

	roles := claims["roles"].(string)
	admin := HasRole(roles, auth.RoleAdmin)

	//admin :=  claims.HasRole(auth.RoleAdmin)
	owner := isRestaurantOwner(restaurant, claims)

	if !owner || !admin {
		return nil, db.ErrForbidden
	}

	// only restaurant owner or admin users can perform menu update

	if menu == nil {
		return r.CreateRestaurantMenu(ctx, um)
	}
	return menu, nil
}

func (r *Repo) CreateRestaurantMenu(ctx context.Context, um UpdateMenu) (*Menu, error) {

	menu, err := r.readMenuByRestaurantDate(ctx, um.RestaurantID, um.Date)
	if err != nil {
		if err != ErrNotFound {
			return nil, errors.Wrapf(err,
				"selecting menu restaurant_id: %s, date: %s",
				um.RestaurantID, um.Date)
		}
	}

	if menu != nil {
		if um.ID == "" {
			um.ID = menu.ID
		}
		return r.updateRestaurantMenu(ctx, um)
	}

	return r.insertRestaurantMenu(ctx, um)
}

func (r *Repo) updateRestaurantMenu(ctx context.Context, um UpdateMenu) (*Menu, error) {

	const qUpdate = `UPDATE menu SET
	    menu =  $1, date = $2
	    WHERE menu_id = $3`

	result, err := r.db.ExecContext(ctx, qUpdate, um.Menu, um.Date, um.ID)
	if err != nil {
		return nil, errors.Wrap(err, "updating menu")
	}

	updateCount, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "updated count")
	}
	if updateCount == 0 {
		return nil, errors.New("no updates done")
	}

	return r.RetrieveMenu(ctx, um.ID)
}

func (r *Repo) insertRestaurantMenu(ctx context.Context, um UpdateMenu) (*Menu, error) {
	if um.ID == "" {
		um.ID = uuid.New().String()
	}
	const qInsert = `INSERT INTO menu 
	(menu_id, restaurant_id, date, menu, votes)
	VALUES ($1, $2, $3, $4, $5)`
	menuResult, err := r.db.ExecContext(ctx, qInsert, um.ID, um.RestaurantID, um.Date, um.Menu, 0)
	if err != nil {
		return nil, errors.Wrap(err, "inserting menu")
	}

	count, err := menuResult.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, errors.New("failed to save new menu")
	}

	menu := Menu{
		ID:           um.ID,
		RestaurantID: um.RestaurantID,
		Date:         um.Date,
		Menu:         um.Menu,
		Votes:        0,
	}
	return &menu, nil
}

func HasRole(roleshas string, roles ...string) bool {
	r := strings.Split(roleshas, " ")
	for _, has := range r {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}
	return false
}
