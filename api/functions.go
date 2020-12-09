package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// LoginEmployee authenticates an employee
func (s *Server) LoginEmployee(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, "sessionCookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Checks request header and methods.
	if r.Header.Get("Content-type") == "application/json" && r.Method == http.MethodPost {
		newCred := &Credentials{}

		// Parse & decode request body into newCred variable
		err := json.NewDecoder(r.Body).Decode(newCred)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Assigns the pointer return of query for hashed password to storedPassword variable.
		storedPassword, err := QueryPassword(s.db, newCred.Username)
		if err != nil {
			http.Error(w, "username or password incorrect", http.StatusUnauthorized)
			return
		}

		// Use bcrypt to compare the hashes of storedPassword (dereferenced) with user input password.
		err = bcrypt.CompareHashAndPassword([]byte(*storedPassword), []byte(newCred.Password))

		if err != nil {
			http.Error(w, "username or password incorrect", http.StatusUnauthorized)
			return
		}

		// Set session cookie values & save
		session.Values["user"] = newCred.Username
		session.Values["auth"] = true
		err = session.Save(r, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}
