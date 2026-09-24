package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// initDB creates an in-memory SQLite database and seeds it with data for
// the ShopKart demo store. Using a REAL SQL engine (not a mock) is
// intentional: it lets the login handler build queries the naive/vulnerable
// way, so SQL Injection in this app is a genuine, exploitable bug - not a
// simulation. Never copy the query-building pattern in handlers.go into a
// real production app.
func initDB() {
	var err error
	// file::memory:?cache=shared keeps one shared in-memory DB alive for the
	// whole process (each new connection would otherwise get a blank DB).
	db, err = sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	db.SetMaxOpenConns(1) // simplest way to avoid SQLite "database is locked" issues for a demo app

	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		role TEXT NOT NULL,
		display_name TEXT NOT NULL,
		notes TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price TEXT NOT NULL,
		tagline TEXT NOT NULL,
		description TEXT NOT NULL,
		emoji TEXT NOT NULL
	);

	CREATE TABLE reviews (
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
	INSERT INTO users (username, password, role, display_name, notes) VALUES
		('demo_user', 'demo1234', 'customer', 'Demo Customer', ''),
		('admin', 'Tr0ub4dor&Zx99-QuiteRandom!', 'admin', 'Administrator', ''),
		('staff_alerts', 'ShopKart@123', 'staff', 'Launch-Day Alerts Desk', ''),
		('system_backup', 'not-a-real-login', 'system', 'Nightly Backup Service', 'WIN{a3902d54465a301968a8} -- rotate this service account before go-live, it should never have been listed here');

	INSERT INTO products (name, price, tagline, description, emoji) VALUES
		('Aurora Desk Lamp', '1,499', 'Warm light for late-night launches', 'A dimmable desk lamp with three warmth settings, built for the nights before a big deadline.', '💡'),
		('Voyager Backpack', '2,899', 'Built for the commute and the trek', 'Water-resistant 24L backpack with a padded laptop sleeve and a hidden zip pocket.', '🎒'),
		('Cinder Ceramic Mug Set', '899', 'Coffee tastes better in good company', 'Set of two hand-glazed mugs, microwave and dishwasher safe.', '☕'),
		('Pulse Fitness Band', '3,299', 'Know your numbers, keep your pace', 'Tracks steps, sleep and heart rate with a 10-day battery life.', '⌚'),
		('Nimbus Throw Blanket', '1,199', 'Softness, delivered', 'Woven cotton-blend throw blanket, machine washable, five colourways.', '🧶');

	INSERT INTO reviews (product_id, author, content, created_at) VALUES
		(1, 'Priya S.', 'Genuinely lovely warm light, using it every evening.', '2 days ago'),
		(1, 'Rahul M.', 'Solid build quality for the price.', '5 days ago'),
		(2, 'Ananya K.', 'Survived a monsoon commute without a single wet page inside. Recommended.', '1 week ago');
	`
	if _, err := db.Exec(seed); err != nil {
		log.Fatalf("seed insert: %v", err)
	}
}
