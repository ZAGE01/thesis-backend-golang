package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	createGameDB(host, user, password, dbname, port)

	connectGameDB := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	var err error
	DB, err = sql.Open("postgres", connectGameDB)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Database is unreachable:", err)
	}

	createTables()
	log.Println("Database connected successfully")
}

func createTables() {
	sqlFile, err := os.ReadFile("database/db.sql")
	if err != nil {
		log.Fatal("Failed to read db.sql:", err)
	}

	if _, err := DB.Exec(string(sqlFile)); err != nil {
		log.Fatal("Failed to execute db.sql:", err)
	}

	log.Println("Tables created successfully")
}

func createGameDB(host, user, password, dbname, port string) {
	// Connect to default postgres db first to create gamedb if needed
	defaultDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable",
		host, user, password, port,
	)

	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		log.Fatal("Failed to connect to default database:", err)
	}
	defer defaultDB.Close()

	// Create gamedb if it doesn't exist
	var exists bool
	err = defaultDB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbname,
	).Scan(&exists)
	if err != nil {
		log.Fatal("Failed to check database existence:", err)
	}

	if !exists {
		_, err = defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			log.Fatal("Failed to create database:", err)
		}
		log.Printf("Database '%s' created successfully", dbname)
	}
}
