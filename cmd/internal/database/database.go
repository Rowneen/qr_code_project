package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"qr_code/internal/config"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

func MustInit() {
	once.Do(func() {
		cfg := config.Get()
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("path error: %s", err)
		}
		dbPath := filepath.Clean(filepath.Join(currentDir, "db", cfg.StoragePath))
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			panic("failed to open database: " + err.Error())
		}

		if err := db.Ping(); err != nil {
			panic("failed to connect to database: " + err.Error())
		}
	})
}

// global db instance
func Get() *sql.DB {
	if db == nil {
		panic("database not initialized. call MustInit()")
	}
	return db
}
