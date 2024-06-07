package api

import (
	"log"

	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func GetBookHighlights(db *db.DB, env *Env, bookID string) []Entry {
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

	return entries
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
