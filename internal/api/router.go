package api

import (
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

	mux.HandleFunc("/", Index(env))
	mux.HandleFunc("GET /hello", Hello(env))
	mux.HandleFunc("GET /table", Table(env, db))
	mux.HandleFunc("GET /import", ImportPage(env))
	mux.HandleFunc("POST /import", ImportFile(env, db))
	mux.HandleFunc("GET /highlights/{id}", GetHighlights(env, db))
	mux.HandleFunc("POST /highlights/edit/{id}", EditHighlight(env, db))
	mux.HandleFunc("DELETE /highlights/edit/{id}", DeleteHighlight(env, db))
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
