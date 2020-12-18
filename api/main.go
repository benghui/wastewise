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
	dbOpt := os.Getenv("DB_OPTIONS")
	dbOValue := os.Getenv("DB_OPT_VALUE")

	// Initialize db connection.
	dbSettings := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s=%s", dbUsername, dbPassword, dbHost, dbPort, dbName, dbOpt, dbOValue)
	db, err := sql.Open("mysql", dbSettings)

	if err != nil {
		log.Fatal(err)
	}

	// Generate authentication & encryption keys using securecookie package.
	authKey := securecookie.GenerateRandomKey(64)
	encryptionKey := securecookie.GenerateRandomKey(32)

	// Initialize session cookiestore with authkey & encryptionkey
	// This is a security measure to prevent access to cookie values
	store := sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)

	// Set max age & httponly options.
	// httponly true prevents Javascript access to cookies and mitigates XSS attack
	// secure true ensures encrypted request over the HTTPS protocol and mitigates MITM attack
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
		Secure:   true,
	}

	// Create instance of server with db connection pool
	// and session cookiestore.
	s := &Server{db: db, store: store}

	// Create new router instance.
	router := mux.NewRouter()

	router.HandleFunc("/api/v1/employees/login", s.LoginEmployee).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/employees/logout", s.LogoutEmployee).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/employees", s.CreateEmployee).Methods(http.MethodPost)

	router.HandleFunc("/api/v1/products", s.GetProducts).Methods(http.MethodGet)

	router.HandleFunc("/api/v1/wastages", s.GetWastages).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/wastages", s.CreateWastage).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/wastages/{id:[0-9]+}", s.GetSingleWastage).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/wastages/{id:[0-9]+}", s.ModifyWastage).Methods(http.MethodPut)

	router.HandleFunc("/api/v1/wastages/reports/month", s.GetWastagesReportMonthly).Methods(http.MethodGet)

	// Add logging & session middleware to router.
	router.Use(LoggingMiddleware, s.SessionMiddleware)

	// Start https server.
	listenAt := fmt.Sprintf(":%s", port)
	fmt.Printf("AHOY! Listening at port%s\n", listenAt)
	log.Fatal(http.ListenAndServeTLS(listenAt, cert, key, router))
}
