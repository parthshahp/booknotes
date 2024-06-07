package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func ExportMarkdown(env *Env, db *db.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env.InfoLog.Println("Serving export markdown")
		id := r.PathValue("id")

		// Get book from db
		query := `
      SELECT 
        b.title, 
        b.created_on AS book_time,
        b.number_of_pages,
        a.authors,
        e.id AS entry_id,
        e.time AS entry_time,
        e.page,
        e.chapter,
        e.text,
        e.note
      FROM books b
      LEFT JOIN
        (SELECT ba.book_id, GROUP_CONCAT(a.name) AS authors
        FROM book_authors ba
        JOIN authors a ON ba.author_id = a.id
        GROUP BY ba.book_id) a ON b.id = a.book_id
      LEFT JOIN
        entries e ON b.id = e.book_id
      WHERE b.id = ?
      ORDER BY e.time DESC;
  `

		rows, err := db.Query(query, id)
		if err != nil {
			env.ErrorLog.Fatalf("Failed to query books: %s", err)
		}
		defer rows.Close()

		type mdNode struct {
			title    string
			bookTime int64
			pages    int
			authors  string
			entry    Entry
		}

		mdNodeList := []mdNode{}

		for rows.Next() {
			curr := mdNode{}

			err = rows.Scan(
				&curr.title,
				&curr.bookTime,
				&curr.pages,
				&curr.authors,
				&curr.entry.ID,
				&curr.entry.Time,
				&curr.entry.Page,
				&curr.entry.Chapter,
				&curr.entry.Text,
				&curr.entry.Note,
			)
			if err != nil {
				env.ErrorLog.Fatalf("Failed to scan rows: %s", err)
			}

			mdNodeList = append(mdNodeList, curr)
		}
		// Conver to markdown format
		markdownContent := []string{}
		markdownContent = append(markdownContent, "# Title: "+mdNodeList[0].title)
		markdownContent = append(markdownContent, "*Authors: "+mdNodeList[0].authors+"*")
		markdownContent = append(
			markdownContent,
			"*Date: "+time.Unix(mdNodeList[0].bookTime, 0).Format("2006-01-02")+"*",
		)
		markdownContent = append(markdownContent, "")

		for i, node := range mdNodeList {
			if i == 0 || node.entry.Chapter != mdNodeList[i-1].entry.Chapter {
				markdownContent = append(markdownContent, "## "+node.entry.Chapter)
			}

			if i == 0 || node.entry.Page != mdNodeList[i-1].entry.Page {
				markdownContent = append(markdownContent, "### Page "+strconv.Itoa(node.entry.Page))
			}

			markdownContent = append(markdownContent, fmt.Sprintf(">%s\n", node.entry.Text))
			if node.entry.Note != "" {
				markdownContent = append(markdownContent, node.entry.Note)
			}
			markdownContent = append(markdownContent, "")
		}

		result := strings.Join(markdownContent, "\n")

		// Export
		w.Header().Set("Content-Disposition", "attachment; filename="+mdNodeList[0].title+".md")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result))
	})
}
