package main

import (
	"database/sql"
	"log"
	_ "net/http"
	"os"
	"path/filepath"
	"qr_code/internal/config"
	handlers "qr_code/internal/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	config.MustLoad()
	cfg := config.Get()

	currentDir, err := os.Getwd()

	if err != nil {
		log.Fatalf("ошибко пути: %s", err)
	}

	dbPath := filepath.Clean(filepath.Join(currentDir, "db", cfg.StoragePath))

	log.Println("Starting application...")
	log.Printf("Config: %+v\n", cfg)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("SQL Open error:", err)
	}
	defer db.Close()

	log.Println("Database opened, pinging...")

	err = db.Ping()
	if err != nil {
		log.Fatal("Ping failed:", err)
	}

	log.Println("SUCCESS: Connected to database!")

	handlers.RegisterHTTPHandlers()
}
