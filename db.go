package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// initDB creates an SQLite database and seeds it with data for
// the ShopKart demo store.
func initDB() {
	var err error
	db, err = sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL,
		display_name TEXT NOT NULL,
		notes TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price TEXT NOT NULL,
		tagline TEXT NOT NULL,
		description TEXT NOT NULL,
		emoji TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS reviews (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER NOT NULL,
		author TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TEXT NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("schema create: %v", err)
	}

	seed := `
	INSERT INTO users (id, username, password, role, display_name, notes) VALUES
		(1, 'demo_user', 'demo1234', 'customer', 'Demo Customer', ''),
		(2, 'admin', 'Tr0ub4dor&Zx99-QuiteRandom!', 'admin', 'Administrator', ''),
		(3, 'staff_alerts', 'ShopKart@123', 'staff', 'Launch-Day Alerts Desk', ''),
		(4, 'system_backup', 'not-a-real-login', 'system', 'Nightly Backup Service', 'WIN{a3902d54465a301968a8} -- rotate this service account before go-live, it should never have been listed here');

	INSERT INTO products (id, name, price, tagline, description, emoji) VALUES
		(1, 'Aurora Desk Lamp', '1,499', 'Warm light for late-night launches', 'A dimmable desk lamp with three warmth settings, built for the nights before a big deadline.', '💡'),
		(2, 'Voyager Backpack', '2,899', 'Built for the commute and the trek', 'Water-resistant 24L backpack with a padded laptop sleeve and a hidden zip pocket.', '🎒'),
		(3, 'Cinder Ceramic Mug Set', '899', 'Coffee tastes better in good company', 'Set of two hand-glazed mugs, microwave and dishwasher safe.', '☕'),
		(4, 'Pulse Fitness Band', '3,299', 'Know your numbers, keep your pace', 'Tracks steps, sleep and heart rate with a 10-day battery life.', '⌚'),
		(5, 'Nimbus Throw Blanket', '1,199', 'Softness, delivered', 'Woven cotton-blend throw blanket, machine washable, five colourways.', '🧶');

	INSERT INTO reviews (id, product_id, author, content, created_at) VALUES
		(1, 1, 'Priya S.', 'Genuinely lovely warm light, using it every evening.', '2 days ago'),
		(2, 1, 'Rahul M.', 'Solid build quality for the price.', '5 days ago'),
		(3, 2, 'Ananya K.', 'Survived a monsoon commute without a single wet page inside. Recommended.', '1 week ago');
	`
	if _, err := db.Exec(seed); err != nil {
		log.Fatalf("seed insert: %v", err)
	}
}
