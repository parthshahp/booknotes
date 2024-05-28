package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func ImportTest() Book {
	// Open JSON file
	jsonFile, err := os.Open("/home/parth/test.json")
	if err != nil {
		log.Fatalf("Failed to open JSON file: %s", err)
	}
	defer jsonFile.Close()

	// Decode JSON data
	var book Book
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&book); err != nil {
		log.Fatalf("Failed to decode JSON: %s", err)
	}

	// Convert epockCreatedOn to time.Time
	book.TimeCreatedOn = time.Unix(book.EpochCreatedOn, 0)
	book.TimeCreatedOn.Format("2006-01-02 15:04:05")

	return book
}

func InsertData(book Book, db *db.DB) {
	// Insert book data
	insertBook := `INSERT INTO books (created_on, number_of_pages, title) VALUES (?, ?, ?)`
	res, err := db.Exec(
		insertBook,
		book.EpochCreatedOn,
		book.NumberOfPages,
		book.Title,
	)
	if err != nil {
		log.Fatalf("Failed to insert book data: %s", err)
	}

	// Get the book_id of the inserted book
	bookID, err := res.LastInsertId()
	if err != nil {
		log.Fatalf("Failed to get last insert id: %s", err)
	}

	// Split authors and insert if they don't exist
	authors := strings.Split(book.Author, "\n")
	for _, author := range authors {
		// Check if author already exists
		var authorID int64
		queryAuthor := `SELECT id FROM authors WHERE name = ?`
		err = db.QueryRow(queryAuthor, author).Scan(&authorID)
		if err != nil && err != sql.ErrNoRows {
			log.Fatalf("Failed to query author: %s", err)
		}

		if err == sql.ErrNoRows {
			// Insert author if not exists
			insertAuthor := `INSERT INTO authors (name) VALUES (?)`
			res, err := db.Exec(insertAuthor, author)
			if err != nil {
				log.Fatalf("Failed to insert author data: %s", err)
			}

			// Get the author_id of the inserted author
			authorID, err = res.LastInsertId()
			if err != nil {
				log.Fatalf("Failed to get last insert id: %s", err)
			}
		}

		// Link book and author
		insertBookAuthor := `INSERT INTO book_authors (book_id, author_id) VALUES (?, ?)`
		if _, err := db.Exec(insertBookAuthor, bookID, authorID); err != nil {
			log.Fatalf("Failed to insert book_author data: %s", err)
		}
	}

	// Insert entries data
	insertEntry := `INSERT INTO entries (book_id, time, page, chapter, text, note) VALUES (?, ?, ?, ?, ?, ?)`
	for _, entry := range book.Entries {
		if _, err := db.Exec(insertEntry, bookID, entry.Time, entry.Page, entry.Chapter, entry.Text, entry.Note); err != nil {
			log.Fatalf("Failed to insert entry data: %s", err)
		}
	}

	fmt.Println("Data inserted successfully")
}
