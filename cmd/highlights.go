package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Entry struct {
	Time    uint64 `json:"time"`
	Page    int    `json:"page"`
	Chapter string `json:"chapter"`
	Text    string `json:"text"`
	Note    string `json:"note"`
}

type Book struct {
	CreatedOn     uint64  `json:"created_on"`
	NumberOfPages int     `json:"number_of_pages"`
	Title         string  `json:"title"`
	Entries       []Entry `json:"entries"`
	Author        string  `json:"author"`
}

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

	return book
}

func InsertData(book Book) {
	// Insert book data
	insertBook := `INSERT INTO books (created_on, number_of_pages, title) VALUES (?, ?, ?)`
	res, err := DB.Exec(
		insertBook,
		book.CreatedOn,
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
		err = DB.QueryRow(queryAuthor, author).Scan(&authorID)
		if err != nil && err != sql.ErrNoRows {
			log.Fatalf("Failed to query author: %s", err)
		}

		if err == sql.ErrNoRows {
			// Insert author if not exists
			insertAuthor := `INSERT INTO authors (name) VALUES (?)`
			res, err := DB.Exec(insertAuthor, author)
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
		if _, err := DB.Exec(insertBookAuthor, bookID, authorID); err != nil {
			log.Fatalf("Failed to insert book_author data: %s", err)
		}
	}

	// Insert entries data
	insertEntry := `INSERT INTO entries (book_id, time, page, chapter, text, note) VALUES (?, ?, ?, ?, ?, ?)`
	for _, entry := range book.Entries {
		if _, err := DB.Exec(insertEntry, bookID, entry.Time, entry.Page, entry.Chapter, entry.Text, entry.Note); err != nil {
			log.Fatalf("Failed to insert entry data: %s", err)
		}
	}

	fmt.Println("Data inserted successfully")
}
