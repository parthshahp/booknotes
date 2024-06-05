package api

import (
	"encoding/json"
	"net/http"

	"github.com/parthshahp/booknotes/internal/db"
	. "github.com/parthshahp/booknotes/internal/types"
)

func ImportJson(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bookImport BookImport

		if err := json.NewDecoder(r.Body).Decode(&bookImport); err != nil {
			http.Error(w, "Unable to parse json", http.StatusBadRequest)
			env.ErrorLog.Println("Error parsing json:", err)
			return
		}

		err := InsertData(bookImport, db, env)
		if err != nil {
			http.Error(w, "Unable to insert data", http.StatusBadRequest)
			env.ErrorLog.Println("Error inserting data:", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
