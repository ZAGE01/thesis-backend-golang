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
	sslmode := os.Getenv("DB_SSLMODE")

	if sslmode == "" {
		sslmode = "disable"
	}

	connectGameDB := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
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
