package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func OpenDB(loc string) (*DB, error) {
	db, err := sql.Open("sqlite3", loc)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db DB) CloseDB() error {
	return db.Close()
}

func (db DB) InitDB() error {
	// Create tables if they don't exist
	createTables := `
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_on INTEGER,
    number_of_pages INTEGER,
		title TEXT
	);
  CREATE TABLE IF NOT EXISTS authors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE
  );
  CREATE TABLE IF NOT EXISTS book_authors (
    book_id INTEGER,
    author_id INTEGER,
    FOREIGN KEY (book_id) REFERENCES books (id),
    FOREIGN KEY (author_id) REFERENCES authors (id),
    PRIMARY KEY (book_id, author_id)
  );
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		book_id INTEGER,
		time INTEGER,
		page INTEGER,
		chapter TEXT,
		text TEXT,
    note TEXT,
		FOREIGN KEY (book_id) REFERENCES books (id)
	);
  CREATE TABLE IF NOT EXISTS book_images (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      book_id INTEGER,
      image BLOB,
      FOREIGN KEY (book_id) REFERENCES books (id)
  );
	`
	if _, err := db.Exec(createTables); err != nil {
		panic(err)
	}

	return nil
}
