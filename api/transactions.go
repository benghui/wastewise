package main

import (
	"database/sql"
	"log"
)


// QueryPassword returns the pointer to hashed password in database for a given username
func QueryPassword(db *sql.DB, username string) (*string, error) {
	// Declare password variable as pointer to string
	var password *string
	row := db.QueryRow("SELECT password from employees where username=?", username)
	err := row.Scan(&password)

	switch err {
	case sql.ErrNoRows:
		log.Println("No rows returned!")
		return nil, err
	case nil:
		return password, nil
	default:
		return nil, err
	}
}
