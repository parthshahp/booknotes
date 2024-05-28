package api

import (
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"

	ui "github.com/parthshahp/booknotes/components"
	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func RoutesInit(env *Env, db *db.DB) http.Handler {
	env.InfoLog.Println("Serving routes")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fs))

	mux.HandleFunc("/", index(env))
	mux.HandleFunc("GET /table", table(env, db))
	mux.HandleFunc("GET /import", importPage(env))
	return logger(mux)
}

func index(env *Env) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving index")
		component := ui.Hello("Parth Shah")
		templ.Handler(component).ServeHTTP(w, r)
	})
}

func table(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		books := GetAllBooks(db, env)
		env.InfoLog.Println("Serving table")
		env.InfoLog.Println(books)
		tableComp := ui.BookTable(books)
		templ.Handler(tableComp).ServeHTTP(w, r)
	})
}

func importPage(env *Env) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving import")
		templ.Handler(ui.Import()).ServeHTTP(w, r)
	})
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
