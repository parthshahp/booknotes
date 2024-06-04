package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
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

func GetBook(db *db.DB, env *Env, bookID string) Book {
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
  WHERE b.id = ?
  GROUP BY b.id
  ORDER BY b.created_on DESC;
  `

	var id int
	var createdOn int64
	var numberOfPages int
	var title string
	var authors string
	var entryCount int
	err := db.QueryRow(query, bookID).
		Scan(&id, &createdOn, &numberOfPages, &title, &authors, &entryCount)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to query book: %s", err)
	}
	authorList := strings.Split(authors, ",")

	env.InfoLog.Println(authors)

	book := Book{
		ID:            id,
		TimeCreatedOn: time.Unix(createdOn, 0),
		NumberOfPages: numberOfPages,
		Title:         title,
		EntryCount:    entryCount,
		Authors:       authorList,
	}
	return book
}

func GetAllBooks(db *db.DB, env *Env, search string) []Book {
	env.InfoLog.Println("Getting all books")
	var rows *sql.Rows
	var err error

	if search == "" {

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

		rows, err = db.Query(query)
		if err != nil {
			log.Fatalf("Failed to query books: %s", err)
		}
	} else {

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
      WHERE b.title LIKE ? or a.authors LIKE ?
      GROUP BY b.id
      ORDER BY b.created_on DESC;
    `

		rows, err = db.Query(query, "%"+search+"%", "%"+search+"%")
		if err != nil {
			log.Fatalf("Failed to query books: %s", err)
		}
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

		book := Book{
			ID:            bookID,
			TimeCreatedOn: time.Unix(createdOn, 0),
			NumberOfPages: numberOfPages,
			Title:         title,
			EntryCount:    entryCount,
			Authors:       authorList,
		}
		books = append(books, book)
	}

	return books
}

func GetBookHighlights(db *db.DB, env *Env, bookID string) (string, []Entry) {
	env.InfoLog.Println("Getting highlights")
	query := `SELECT id, time, page, chapter, text, note FROM entries WHERE book_id = ? ORDER BY page DESC;`
	queryName := `SELECT title FROM books WHERE id = ?;`
	rows, err := db.Query(query, bookID)
	if err != nil {
		log.Fatalf("Failed to query entries: %s", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.Time, &entry.Page, &entry.Chapter, &entry.Text, &entry.Note); err != nil {
			log.Fatalf("Failed to scan entry: %s", err)
		}
		entries = append(entries, entry)
	}

	var bookName string
	rows, err = db.Query(queryName, bookID)
	if err != nil {
		log.Fatalf("Failed to query name: %s", err)
	}
	defer rows.Close()

	rows.Next()
	if err := rows.Scan(&bookName); err != nil {
		log.Fatalf("Failed to scan name: %s", err)
	}

	return bookName, entries
}

func UpdateHighlight(db *db.DB, env *Env, highlight Entry) Entry {
	query := `UPDATE entries SET page = ?, chapter = ?, text = ?, note = ? WHERE id = ?;`
	stmt, err := db.Prepare(query)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to prepare query: %s", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(highlight.Page, highlight.Chapter, highlight.Text, highlight.Note, highlight.ID); err != nil {
		env.ErrorLog.Fatalf("Failed to update highlight: %s", err)
	}

	// Return the updated highlight from the db
	query = `SELECT id, time, page, chapter, text, note FROM entries WHERE id = ?;`
	rows, err := db.Query(query, highlight.ID)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to query entry: %s", err)
	}
	defer rows.Close()

	var updatedEntry Entry
	for rows.Next() {
		if err := rows.Scan(&updatedEntry.ID, &updatedEntry.Time, &updatedEntry.Page, &updatedEntry.Chapter, &updatedEntry.Text, &updatedEntry.Note); err != nil {
			env.ErrorLog.Fatalf("Failed to scan entry: %s", err)
		}
	}
	return updatedEntry
}

func UpdateBook(db *db.DB, env *Env, title, id string, authors []string) {
	query := `UPDATE books SET title = ? WHERE id = ?;`
	stmt, err := db.Prepare(query)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to prepare query: %s", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(title, id); err != nil {
		env.ErrorLog.Fatalf("Failed to update book: %s", err)
	}

	// Delete the original authors
	query = `DELETE FROM book_authors WHERE book_id = ?;`
	_, err = db.Exec(query, id)
	if err != nil {
		env.ErrorLog.Fatalf("Failed to delete authors: %s", err)
	}

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
		if _, err := db.Exec(insertBookAuthor, id, authorID); err != nil {
			env.ErrorLog.Fatalf("Failed to insert book_author data: %s", err)
		}
	}
}
