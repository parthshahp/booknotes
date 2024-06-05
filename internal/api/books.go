package api

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

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
