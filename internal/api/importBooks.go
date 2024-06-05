package api

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func InsertData(book BookImport, db *db.DB, env *Env) {
	env.InfoLog.Println("Inserting data")
	insertBook := `INSERT INTO books (created_on, number_of_pages, title) VALUES (?, ?, ?)`
	res, err := db.Exec(
		insertBook,
		book.EpochCreatedOn,
		book.NumberOfPages,
		book.Title,
	)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to insert book data: %s", err)
	}

	// Get the book_id of the inserted book
	bookID, err := res.LastInsertId()
	if err != nil {
		env.ErrorLog.Fatalf("Failed to get last insert id: %s", err)
	}

	// Split authors and insert if they don't exist
	authors := strings.Split(book.Author, "\n")
	for _, author := range authors {
		// Check if author already exists
		var authorID int64
		queryAuthor := `SELECT id FROM authors WHERE name = ?`
		err = db.QueryRow(queryAuthor, author).Scan(&authorID)
		if err != nil && err != sql.ErrNoRows {
			env.ErrorLog.Fatalf("Failed to query author: %s", err)
		}

		if err == sql.ErrNoRows {
			// Insert author if not exists
			insertAuthor := `INSERT INTO authors (name) VALUES (?)`
			res, err := db.Exec(insertAuthor, author)
			if err != nil {
				env.ErrorLog.Fatalf("Failed to insert author data: %s", err)
			}

			// Get the author_id of the inserted author
			authorID, err = res.LastInsertId()
			if err != nil {
				env.ErrorLog.Fatalf("Failed to get last insert id: %s", err)
			}
		}

		// Link book and author
		insertBookAuthor := `INSERT INTO book_authors (book_id, author_id) VALUES (?, ?)`
		if _, err := db.Exec(insertBookAuthor, bookID, authorID); err != nil {
			env.ErrorLog.Fatalf("Failed to insert book_author data: %s", err)
		}
	}

	// Insert entries data
	insertEntry := `INSERT INTO entries (book_id, time, page, chapter, text, note) VALUES (?, ?, ?, ?, ?, ?)`
	for _, entry := range book.Entries {
		if _, err := db.Exec(insertEntry, bookID, entry.Time, entry.Page, entry.Chapter, entry.Text, entry.Note); err != nil {
			env.ErrorLog.Fatalf("Failed to insert entry data: %s", err)
		}
	}

	env.InfoLog.Println("Data inserted successfully")
}

func TestQuery(db *db.DB, env *Env) {
	env.InfoLog.Println("Testing query")
	rows, err := db.Query("SELECT title FROM books")
	if err != nil {
		log.Fatalf("Failed to query books: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bookTitle string
		rows.Scan(&bookTitle)
		env.InfoLog.Println(bookTitle)
		env.InfoLog.Println("Should be a book")
	}
}
