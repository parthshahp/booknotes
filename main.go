package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
)

type Env struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func main() {
	addr := flag.String("addr", ":3000", "The addr of the application.")
	flag.Parse()

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
		Addr:     *addr,
		Handler:  handler,
		ErrorLog: errorLog,
	}

	infoLog.Printf("Starting server on %s", *addr)
	errorLog.Fatal(server.ListenAndServe())
}

func routesInit(env *Env) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index(env))
	return logger(mux)
}

func index(env *Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := InitDB(env)
		if err != nil {
			env.errorLog.Fatal(err)
		}
		env.infoLog.Println("DB initialized")

		book := ImportTest()
		env.infoLog.Println("Book imported")

		InsertData(book)
		env.infoLog.Println("Data inserted")

		w.Write([]byte("Hello, World!"))
	}
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
