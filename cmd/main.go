package main

import (
	"fmt"
	"log"
	"postgresDB/config"
	"postgresDB/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	// load env variables
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("errorload .env file")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	db, err := database.NewConnection(cfg.DB)
	if err != nil {
		fmt.Println("DB Connection Error:", err)
		return
	}
	defer db.Close()
	fmt.Println("DB pool initialized successfully")
}
