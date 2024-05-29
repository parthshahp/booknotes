package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/parthshahp/booknotes/internal/api"
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
	db.InitDB()

	c := cors.AllowAll()
	handler := c.Handler(api.RoutesInit(&env, db))
	server := http.Server{
		Addr:     addr,
		Handler:  handler,
		ErrorLog: errorLog,
	}

	infoLog.Printf("Starting server on %s", addr)
	errorLog.Fatal(server.ListenAndServe())
}
