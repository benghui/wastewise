package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

// Server struct holding database connection pool.
type Server struct {
	db    *sql.DB
	store *sessions.CookieStore
}

func init() {
	// Load variables in .env file using godotenv package.
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

	// Generate authentication & encryption keys using securecookie package.
	authKey := securecookie.GenerateRandomKey(64)
	encryptionKey := securecookie.GenerateRandomKey(32)

	// Initialize session cookie store.
	store := sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)

	// Set max age & httponly options
	store.Options = &sessions.Options{
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	// Create instance of server with db connection pool.
	s := &Server{db: db, store: store}

	// Create new router instance.
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/employees/login", s.LoginEmployee).Methods(http.MethodPost)

	// Start https server
	listenAt := fmt.Sprintf(":%s", port)
	fmt.Printf("AHOY! Listening at port%s\n", listenAt)
	log.Fatal(http.ListenAndServeTLS(listenAt, cert, key, router))
}
