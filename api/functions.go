package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// LoginEmployee authenticates an employee
func (s *Server) LoginEmployee(w http.ResponseWriter, r *http.Request) {
	// Loads the session data from cookiestore.
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Assigns the pointer return of query for hashed password to storedPassword variable.
		storedPassword, employeeID, err := QueryPassword(s.db, newCred.Username)

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
		session.Values["ID"] = employeeID
		session.Values["auth"] = true
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/api/v1/wastages", http.StatusFound)

	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// LogoutEmployee revokes authentication.
func (s *Server) LogoutEmployee(w http.ResponseWriter, r *http.Request) {
	// Loads the session data from cookiestore.
	session, err := s.store.Get(r, "sessionCookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["user"] = nil
	session.Values["ID"] = nil
	session.Values["auth"] = nil
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetWastage handles query to wastage data.
// Default is last 7 days from current day.
func (s *Server) GetWastages(w http.ResponseWriter, r *http.Request) {
	// Loads the session data from cookiestore.
	session, err := s.store.Get(r, "sessionCookie")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check that employee is authenticated. Otherwise redirect to login.
	if session.Values["auth"] != true {
		http.Redirect(w, r, "/api/v1/employee/login", http.StatusUnauthorized)
		return
	}

	wastageResults, err := QueryWastage(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header to parse response as JSON.
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(wastageResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
}
