package db

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

func Init(dataSourceName string) error {
	var err error
	once.Do(func() {
		db, err = sql.Open("sqlite3", dataSourceName)
		if err == nil {
			// Create tables if not exist
			_, err = db.Exec(`
				CREATE TABLE IF NOT EXISTS trade_links (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					url TEXT NOT NULL,
					description TEXT,
				    selected INTEGER NOT NULL
				);
				CREATE TABLE IF NOT EXISTS settings (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					key TEXT NOT NULL UNIQUE,
					value TEXT
				);
				CREATE TABLE IF NOT EXISTS live_search_settings (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT NOT NULL,
					enabled INTEGER NOT NULL
				);
			`)
		}
	})
	return err
}

func GetDB() *sql.DB {
	return db
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
