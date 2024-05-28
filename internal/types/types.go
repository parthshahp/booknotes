package types

import (
	"log"
	"time"
)

type Env struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

type Entry struct {
	Time    uint64 `json:"time"`
	Page    int    `json:"page"`
	Chapter string `json:"chapter"`
	Text    string `json:"text"`
	Note    string `json:"note"`
}

type BookImport struct {
	EpochCreatedOn int64   `json:"created_on"`
	NumberOfPages  int     `json:"number_of_pages"`
	Title          string  `json:"title"`
	Entries        []Entry `json:"entries"`
	Author         string  `json:"author"`
}

type Book struct {
	ID            int
	TimeCreatedOn time.Time
	NumberOfPages int
	Title         string
	EntryCount    int
	Authors       []string
}
