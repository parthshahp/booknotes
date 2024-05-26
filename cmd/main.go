package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
	ui "github.com/parthshahp/booknotes/components"
	"github.com/rs/cors"
)

type Env struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	addr := os.Getenv("LISTEN_ADDR")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	env := Env{
		infoLog:  infoLog,
		errorLog: errorLog,
	}

	err := OpenDB()
	if err != nil {
		errorLog.Fatal(err)
	}
	defer CloseDB()

	c := cors.AllowAll()
	handler := c.Handler(routesInit(&env))
	server := http.Server{
		Addr:     addr,
		Handler:  handler,
		ErrorLog: errorLog,
	}

	infoLog.Printf("Starting server on %s", addr)
	errorLog.Fatal(server.ListenAndServe())
}

func routesInit(env *Env) http.Handler {
	env.infoLog.Println("Serving routes")
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fs))

	mux.Handle("/", index(env))
	return logger(mux)
}

func index(env *Env) http.Handler {
	env.infoLog.Println("Serving index")
	component := ui.Hello("Parth Shah")
	return templ.Handler(component)
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
