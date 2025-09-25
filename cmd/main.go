package main

import (
	"log"
	"qr_code/internal/config"
	"qr_code/internal/database"
	"qr_code/internal/handlers"
)

func main() {
	log.Println("Starting application...")

	config.MustLoad()
	database.MustInit()

	log.Println("SUCCESS: Database initialized!")

	handlers.RegisterHTTPHandlers()
}
