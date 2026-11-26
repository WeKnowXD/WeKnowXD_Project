package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dsn string) {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}
}

func InitFTS() {
	_, err := DB.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pages_fts USING fts5(title, content, content='pages', content_rowid='rowid')")
	if err != nil {
		log.Println("FTS5 setup failed:", err)
		return
	}
	if _, err = DB.Exec("INSERT INTO pages_fts(pages_fts) VALUES('rebuild')"); err != nil {
		log.Println("FTS5 rebuild failed:", err)
	}
}
