package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"

	ui "github.com/parthshahp/booknotes/components"
	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func Index(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving index")
		// templ.Handler(ui.Page()).ServeHTTP(w, r)
		books := GetAllBooks(db, env, "")
		templ.Handler(ui.Page(books)).ServeHTTP(w, r)
	})
}

func Table(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		books := GetAllBooks(db, env, "")
		env.InfoLog.Println("Serving table")
		templ.Handler(ui.BookTable(books)).ServeHTTP(w, r)
	})
}

func ImportPage(env *Env) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving import")
		templ.Handler(ui.Import()).ServeHTTP(w, r)
	})
}

func ImportFile(env *Env, db *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving import")

		// Parse the multipart form
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			http.Error(w, "Unable to parse multipart form", http.StatusBadRequest)
			env.ErrorLog.Println("Error parsing multipart form:", err)
			return
		}

		files := r.MultipartForm.File["file"]
		if len(files) == 0 {
			http.Error(w, "No files uploaded", http.StatusBadRequest)
			env.ErrorLog.Println("No files found in the request")
			return
		}

		for _, f := range files {
			file, err := f.Open()
			if err != nil {
				http.Error(w, "Unable to open file", http.StatusInternalServerError)
				env.ErrorLog.Println("Error opening file:", err)
				return
			}
			defer file.Close()

			var book BookImport

			data, err := io.ReadAll(file)
			if err != nil {
				env.ErrorLog.Fatalf("Failed to read file: %s", err)
			}

			err = json.Unmarshal(data, &book)
			if err != nil {
				env.ErrorLog.Fatalf("Failed to unmarshal JSON: %s", err)
			}

			InsertData(book, db, env)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("File uploaded successfully"))
	}
}

func GetHighlights(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookID := r.PathValue("id")
		env.InfoLog.Println("Serving highlights for book", bookID)
		if bookID == "" {
			http.Error(w, "No book ID provided", http.StatusBadRequest)
			env.ErrorLog.Println("No book ID provided")
			return
		}
		entries := GetBookHighlights(db, env, bookID)
		book := GetBook(db, env, bookID)
		templ.Handler(ui.HighlightsPage(book, entries, bookID)).ServeHTTP(w, r)
	})
}

func EditHighlight(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving edit highlight")
		pathID := r.PathValue("id")
		// Create Entry
		if pathID == "" {
			http.Error(w, "No highlight ID provided", http.StatusBadRequest)
			env.ErrorLog.Println("No highlight ID provided")
			return
		}
		// Convert highlightID to int from string
		highlightID, err := strconv.Atoi(pathID)
		if err != nil {
			http.Error(w, "Invalid highlight ID provided", http.StatusBadRequest)
			env.ErrorLog.Println("Invalid highlight ID provided")
			return
		}

		// Parse the form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			env.ErrorLog.Println("Error parsing form:", err)
			return
		}
		updatedEntry := Entry{ID: highlightID}
		updatedEntry.Page, err = strconv.Atoi(r.FormValue("page"))
		if err != nil {
			http.Error(w, "Invalid page number provided", http.StatusBadRequest)
			env.ErrorLog.Println("Invalid page number provided")
			return
		}
		updatedEntry.Chapter = r.FormValue("chapter")
		updatedEntry.Text = r.FormValue("text")
		updatedEntry.Note = r.FormValue("note")

		updatedEntry = UpdateHighlight(db, env, updatedEntry)

		templ.Handler(ui.Highlight(fmt.Sprintf("%d", updatedEntry.ID), updatedEntry.Chapter, updatedEntry.Text, updatedEntry.Note, fmt.Sprintf("%d", updatedEntry.Page), time.Unix(updatedEntry.Time, 0).Format("2006-01-02"))).
			ServeHTTP(w, r)
	})
}

func DeleteHighlight(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving delete highlight")
		pathID := r.PathValue("id")
		if pathID == "" {
			http.Error(w, "No highlight ID provided", http.StatusBadRequest)
			env.ErrorLog.Println("No highlight ID provided")
			return
		}

		query := `DELETE FROM entries WHERE id = ?;`
		stmt, err := db.Prepare(query)
		if err != nil {
			env.ErrorLog.Fatalf("Failed to prepare query: %s", err)
		}
		defer stmt.Close()

		if _, err := stmt.Exec(pathID); err != nil {
			env.ErrorLog.Fatalf("Failed to delete highlight: %s", err)
		}
	})
}

func EditBook(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving edit book")
		pathID := r.PathValue("id")
		if pathID == "" {
			http.Error(w, "No book ID provided", http.StatusBadRequest)
			env.ErrorLog.Println("No book ID provided")
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			env.ErrorLog.Println("Error parsing form:", err)
			return
		}

		var authors []string
		// Split authors by comma
		for _, author := range strings.Split(r.FormValue("author"), ",") {
			authors = append(authors, strings.TrimSpace(author))
		}

		UpdateBook(db, env, r.FormValue("title"), pathID, authors)
		book := GetBook(db, env, pathID)
		templ.Handler(
			ui.BookTableEntry(
				book.Title,
				strings.Join(book.Authors, ", "),
				book.TimeCreatedOn.Format("2006-01-02"),
				strconv.Itoa(book.EntryCount),
				fmt.Sprintf("%d", book.ID),
			),
		).
			ServeHTTP(w, r)
	})
}

func DeleteBook(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving delete book")
		pathID := r.PathValue("id")
		if pathID == "" {
			http.Error(w, "No book ID provided", http.StatusBadRequest)
			env.ErrorLog.Println("No book ID provided")
			return
		}

		query := `DELETE FROM books WHERE id = ?;`
		stmt, err := db.Prepare(query)
		if err != nil {
			env.ErrorLog.Fatalf("Failed to prepare query: %s", err)
		}
		defer stmt.Close()

		if _, err := stmt.Exec(pathID); err != nil {
			env.ErrorLog.Fatalf("Failed to delete book: %s", err)
		}

		query = `DELETE FROM book_authors WHERE book_id = ?;`
		stmt, err = db.Prepare(query)
		if err != nil {
			env.ErrorLog.Fatalf("Failed to prepare query: %s", err)
		}
		defer stmt.Close()

		if _, err := stmt.Exec(pathID); err != nil {
			env.ErrorLog.Fatalf("Failed to delete authors: %s", err)
		}
	})
}

func SearchBookTable(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving search book table")
		query := r.FormValue("search")
		env.InfoLog.Println("Search query:", query)
		books := GetAllBooks(db, env, query)
		templ.Handler(ui.BookTableTable(books)).ServeHTTP(w, r)
	})
}

func SearchHighlightsPage(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving highlights")
		templ.Handler(ui.HighlightsSearch()).ServeHTTP(w, r)
	})
}

func SearchHighlights(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving search highlights")
		query := r.FormValue("search")
		env.InfoLog.Println("Search query:", query)
		highlights := SearchAllHighlights(db, env, query)
		env.InfoLog.Println("Found", len(highlights), "highlights")
		templ.Handler(ui.HighlightResults(highlights)).ServeHTTP(w, r)
	})
}
