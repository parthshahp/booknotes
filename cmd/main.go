package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	ui "github.com/parthshahp/booknotes/components"
	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	addr := os.Getenv("LISTEN_ADDR")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	env := Env{
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}

	db, err := db.OpenDB()
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.CloseDB()

	c := cors.AllowAll()
	handler := c.Handler(routesInit(&env, db))
	server := http.Server{
		Addr:     addr,
		Handler:  handler,
		ErrorLog: errorLog,
	}

	infoLog.Printf("Starting server on %s", addr)
	errorLog.Fatal(server.ListenAndServe())
}

func routesInit(env *Env, db *db.DB) http.Handler {
	env.InfoLog.Println("Serving routes")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fs))

	mux.Handle("/", index(env))
	mux.Handle("GET /table", table(env, db))
	return logger(mux)
}

func index(env *Env) http.Handler {
	env.InfoLog.Println("Serving index")
	component := ui.Hello("Parth Shah")
	return templ.Handler(component)
}

func table(env *Env, db *db.DB) http.Handler {
	temp := Book{
		Title:          "Test",
		Author:         "Parth Shah",
		EpochCreatedOn: time.Now().Unix(),
		NumberOfPages:  100,
		Entries:        []Entry{},
	}
	temp.TimeCreatedOn = time.Unix(temp.EpochCreatedOn, 0)
	env.InfoLog.Println("Serving table")

	// component := ui.BookTableEntry(
	// 	temp.Title,
	// 	temp.Author,
	// 	temp.TimeCreatedOn.Format("2006-01-02 15:04:05"),
	// 	strconv.Itoa(temp.NumberOfPages),
	// )
	var entries []Book
	entries = append(entries, temp)

	tableComp := ui.BookTable(entries)
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
