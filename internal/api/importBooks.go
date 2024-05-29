package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func ImportHighlightData(f multipart.File, env *Env, db *db.DB) error {
	var book BookImport

	data, err := io.ReadAll(f)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to read file: %s", err)
	}

	err = json.Unmarshal(data, &book)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to unmarshal JSON: %s", err)
	}

	InsertData(book, db, env)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to insert data: %s", err)
	}

	return nil
}

func ImportTest() BookImport {
	// Open JSON file
	jsonFile, err := os.Open("/home/parth/test.json")
	if err != nil {
		log.Fatalf("Failed to open JSON file: %s", err)
	}
	defer jsonFile.Close()

	// Decode JSON data
	var book BookImport
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&book); err != nil {
		log.Fatalf("Failed to decode JSON: %s", err)
	}
	return book
}

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

func GetAllBooks(db *db.DB, env *Env) []Book {
	env.InfoLog.Println("Getting all books")
	// TestQuery(db, env)
	query := `
  SELECT 
    b.id, 
    b.created_on,
    b.number_of_pages,
    b.title, 
    a.authors,
    COUNT(e.id) AS entry_count
  FROM books b
  LEFT JOIN
    (SELECT ba.book_id, GROUP_CONCAT(a.name) AS authors
    FROM book_authors ba
    JOIN authors a ON ba.author_id = a.id
    GROUP BY ba.book_id) a ON b.id = a.book_id
  LEFT JOIN
    entries e ON b.id = e.book_id
  GROUP BY b.id
  ORDER BY b.created_on DESC;
  `

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query books: %s", err)
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var bookID int
		var createdOn int64
		var numberOfPages int
		var title string
		var authors string
		var entryCount int

		if err := rows.Scan(&bookID, &createdOn, &numberOfPages, &title, &authors, &entryCount); err != nil {
			log.Fatalf("Failed to scan book: %s", err)
		}

		authorList := strings.Split(authors, ",")

		env.InfoLog.Println(authors)

		book := Book{
			ID:            bookID,
			TimeCreatedOn: time.Unix(createdOn, 0),
			NumberOfPages: numberOfPages,
			Title:         title,
			EntryCount:    entryCount,
			Authors:       authorList,
		}
		env.InfoLog.Println(book.Title)
		books = append(books, book)
	}

	return books
}

func GetBookHighlights(db *db.DB, env *Env, bookID string) []Entry {
	env.InfoLog.Println("Getting highlights")
	query := `SELECT time, page, chapter, text, note FROM entries WHERE book_id = ? ORDER BY page DESC;`
	rows, err := db.Query(query, bookID)
	if err != nil {
		log.Fatalf("Failed to query entries: %s", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.Time, &entry.Page, &entry.Chapter, &entry.Text, &entry.Note); err != nil {
			log.Fatalf("Failed to scan entry: %s", err)
		}
		entries = append(entries, entry)
	}
	return entries
}
