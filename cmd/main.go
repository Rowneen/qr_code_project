package main

import (
	"fmt"
	_ "net/http"
	"qr_code/internal/config"
	"log"
	"database/sql"
)


func main() {
	cfg := config.MustLoad()
	db, err := sql.Open("sqlite3", "./db/sqlite.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()



	log.Println("Successfully connected to database!")

	fmt.Println(cfg)
	


}
