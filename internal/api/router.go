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

	mux.Handle("/", index(env, db))
	mux.Handle("GET /table", table(env, db))
	mux.Handle("GET /import", importPage(env))
	return logger(mux)
}

func index(env *Env, db *db.DB) http.Handler {
	env.InfoLog.Println("Serving index")
	db.InitDB()
	component := ui.Hello("Parth Shah")
	return templ.Handler(component)
}

func table(env *Env, db *db.DB) http.Handler {
	books := GetAllBooks(db, env)
	env.InfoLog.Println("Serving table")
	env.InfoLog.Println(books)
	tableComp := ui.BookTable(books)
	return templ.Handler(tableComp)
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

func importPage(env *Env) http.Handler {
	env.InfoLog.Println("Serving import")
	return templ.Handler(ui.Import())
}
