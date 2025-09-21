package db

import (
	"database/sql"
	"embed"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

var (
	db   *sql.DB
	once sync.Once
)

func Init(dataSourceName string, migrationsFS embed.FS) error {
	var err error
	once.Do(func() {
		db, err = sql.Open("sqlite3", dataSourceName)
		if err != nil {
			return
		}

		// Configurar goose para usar el provider sqlite3
		goose.SetBaseFS(migrationsFS)

		// Ejecutar las migraciones
		if err = goose.SetDialect("sqlite3"); err != nil {
			return
		}

		if err = goose.Up(db, "migrations"); err != nil {
			return
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
