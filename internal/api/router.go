package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func RoutesInit(env *Env, db *db.DB) http.Handler {
	env.InfoLog.Println("Serving routes")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fs))

	mux.HandleFunc("/", Index(env, db))
	mux.HandleFunc("GET /table", Table(env, db))
	mux.HandleFunc("POST /table/search", SearchBookTable(env, db))
	mux.HandleFunc("GET /import", ImportPage(env))
	mux.HandleFunc("POST /import/file", ImportFile(env, db))
	mux.HandleFunc("POST /import/json", ImportJson(env, db))

	mux.HandleFunc("POST /book/{id}", EditBook(env, db))
	mux.HandleFunc("DELETE /book/{id}", DeleteBook(env, db))
	mux.HandleFunc("GET /book/{id}/highlights", GetHighlights(env, db))

	mux.HandleFunc("GET /highlights", SearchHighlightsPage(env, db))
	mux.HandleFunc("POST /highlights/search", SearchHighlights(env, db))
	mux.HandleFunc("POST /highlights/edit/{id}", EditHighlight(env, db))
	mux.HandleFunc("DELETE /highlights/edit/{id}", DeleteHighlight(env, db))
	mux.HandleFunc("GET /handleExport/{type}/{id}", Export(env))
	mux.HandleFunc("GET /export/markdown/{id}", ExportMarkdown(env, db))

	return logger(mux)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()

		log.Printf(
			"[%s] %s %s %s",
			end.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			end.Sub(start),
		)
	})
}

func Export(env *Env) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving export")
		exportType := r.PathValue("type")
		id := r.PathValue("id")

		// Redirect to the export page
		w.Header().Add("HX-Redirect", fmt.Sprintf("/export/%s/%s", exportType, id))
		w.WriteHeader(http.StatusSeeOther)
	})
}
