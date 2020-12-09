package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Server struct holding database connection pool.
type Server struct {
	db *sql.DB
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// Read environment variable from .env file.
	port := os.Getenv("PORT")
	cert := os.Getenv("CERT")
	key := os.Getenv("KEY")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// Initialize db connection.
	dbSettings := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUsername, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dbSettings)

	if err != nil {
		log.Fatal(err)
	}

	// Create instance of server with db connection pool.
	s := &Server{db: db}

	// Create new router instance.
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/employees/login", s.LoginEmployee).Methods(http.MethodPost)

	// Start https server
	listenAt := fmt.Sprintf(":%s", port)
	fmt.Printf("AHOY! Listening at port%s\n", listenAt)
	log.Fatal(http.ListenAndServeTLS(listenAt, cert, key, router))
}
