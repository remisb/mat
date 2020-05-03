package schema

import "github.com/jmoiron/sqlx"

func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		return err
	}
	return tx.Commit()
}

const seeds = `

INSERT INTO restaurant (restaurant_id, name, address, owner_user_id, date_created, date_updated) VALUES
  ('0ce90028-69cb-4e9c-9af0-7bbada50d5b6', 'Paikis', 'A. Smetonos g. 5, Vilnius 01115', '5cf37266-3473-4006-984f-9325122678b7', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
  ('71b8fb90-24eb-4012-9048-3ba210aac0f6', 'Seeet Root', 'Užupio g. 22, Vilnius 01203', '5cf37266-3473-4006-984f-9325122678b7', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
  ('2df32931-3072-4d11-8109-d1f0988c26b3', 'Lauro lapas', 'Pamėnkalnio g. 24, Vilnius 01114', '5cf37266-3473-4006-984f-9325122678b7', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
  ('8800c4d0-0219-49d5-9eb0-db457ee015e5', 'Mykolo 4', 'Šv. Mykolo g. 4, Vilnius 01124', '5cf37266-3473-4006-984f-9325122678b7', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
  ('5828612a-1f8a-403c-b6d1-6cb66fbf0c66', 'Lokys', 'Stiklių g. 10, Vilnius 01131', '5cf37266-3473-4006-984f-9325122678b7', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
  ON CONFLICT DO NOTHING;

INSERT INTO menu (restaurant_id, date, menu, votes) VALUES
	('5828612a-1f8a-403c-b6d1-6cb66fbf0c66', '2020-03-01 00:00:00', 'Lokys menu for 2020-03-01', 0),
	('5828612a-1f8a-403c-b6d1-6cb66fbf0c66', '2020-03-02 00:00:00', 'Lokys menu for 2020-03-02', 0)
	ON CONFLICT DO NOTHING;

-- Create admin and regular User with password "gophers"
INSERT INTO users (user_id, name, email, roles, password_hash, date_created, date_updated) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', 'admin@example.com', '{ADMIN,USER}', '$2a$10$1ggfMVZV6Js0ybvJufLRUOWHS5f6KneuP0XwwHpJ8L8ipdry9f2/a', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', 'user@example.com', '{USER}', '$2a$10$9/XASPKBbJKVfCAZKDH.UuhsuALDr5vVm6VrYA9VFR8rccK86C1hW', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
	ON CONFLICT DO NOTHING;
`
