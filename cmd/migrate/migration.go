package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"postgresDB/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Database connection string
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
		cfg.DB.SSLMode,
	)

	// Migration source URL (file://path/to/migrations)
	sourceURL := "file://migrations"

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}
	defer m.Close()

	// Parse command
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [command]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  up          Apply all up migrations\n")
		fmt.Fprintf(os.Stderr, "  down        Rollback one migration\n")
		fmt.Fprintf(os.Stderr, "  force <v>   Force set version to <v>\n")
		fmt.Fprintf(os.Stderr, "  version     Print current migration version\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run up migrations: %v", err)
		}
		log.Println("Migrations applied successfully!")

	case "down":
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run down migration: %v", err)
		}
		log.Println("Migration rolled back successfully!")

	case "force":
		if flag.NArg() < 2 {
			log.Fatal("Usage: force <version>")
		}
		version := -1
		fmt.Sscanf(flag.Arg(1), "%d", &version)
		if version == -1 {
			log.Fatal("Invalid version")
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force version: %v", err)
		}
		log.Printf("Forced version to %d\n", version)

	case "version":
		version, dirty, err := m.Version()
		if err != nil && err != migrate.ErrNilVersion {
			log.Fatalf("Failed to get version: %v", err)
		}
		if err == migrate.ErrNilVersion {
			log.Println("No migrations applied yet.")
		} else {
			log.Printf("Current version: %d (Dirty: %v)\n", version, dirty)
		}

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
