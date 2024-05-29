package api

import (
	"net/http"

	"github.com/a-h/templ"

	ui "github.com/parthshahp/booknotes/components"
	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func Index(env *Env) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving index")
		templ.Handler(ui.Page()).ServeHTTP(w, r)
	})
}

func Hello(env *Env) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving hello")
		templ.Handler(ui.Hello("Hi there!")).ServeHTTP(w, r)
	})
}

func Table(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		books := GetAllBooks(db, env)
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

			ImportHighlightData(file, env, db)
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
		bookName, entries := GetBookHighlights(db, env, bookID)
		templ.Handler(ui.HighlightsPage(bookName, entries)).ServeHTTP(w, r)
	})
}
