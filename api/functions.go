package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

	// Checks that request content type matches.
	if r.Header.Get("Content-type") == "application/json" {
		newCred := &Credentials{}

		// Parse & decode request body into newCred variable
		err := json.NewDecoder(r.Body).Decode(newCred)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Assigns the pointer return of query for hashed password to storedPassword variable.
		storedPassword, employeeID, role, err := QueryPassword(s.db, newCred.Username)

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
		session.Values["role"] = role
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// http.Redirect(w, r, "/api/v1/wastages", http.StatusFound)

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

// GetProducts handles query to all products data.
// It is a protected resource requiring authentication value from sessions.
func (s *Server) GetProducts(w http.ResponseWriter, r *http.Request) {
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

	productResults, err := QueryProducts(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header to parse response as JSON.
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(productResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
}

// GetWastages handles query to wastage data.
// It is a protected resource requiring authentication value from sessions.
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

// CreateWastage handles inserting new wastage entry.
func (s *Server) CreateWastage(w http.ResponseWriter, r *http.Request) {
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

	if r.Header.Get("Content-type") == "application/json" {
		newWastageForm := &WastageForm{}

		// Locks & defer Unlock to prevent race conditions
		newWastageForm.Mu.Lock()
		defer newWastageForm.Mu.Unlock()

		err := json.NewDecoder(r.Body).Decode(newWastageForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Read employee ID from session values.
		// Type assert to int.
		employeeID := session.Values["ID"].(int)
		err = CreateWastageEntry(
			s.db,
			newWastageForm.WastageDate,
			newWastageForm.WastageQuantity,
			newWastageForm.WastageReason,
			newWastageForm.ProductID,
			employeeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// GetSingleWastage handles query for wastage entry.
func (s *Server) GetSingleWastage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)["id"]
	id, err := strconv.Atoi(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wastage, err := QuerySingleWastage(s.db, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(wastage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
}

// ModifyWastage handles editing of one wastage record.
func (s *Server) ModifyWastage(w http.ResponseWriter, r *http.Request) {
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

	// Checks that request content type matches.
	if r.Header.Get("Content-type") == "application/json" {
		params := mux.Vars(r)["id"]
		id, err := strconv.Atoi(params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		editWastageForm := &WastageForm{}

		// Locks & defer Unlock to prevent race conditions
		editWastageForm.Mu.Lock()
		defer editWastageForm.Mu.Unlock()

		err = json.NewDecoder(r.Body).Decode(editWastageForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Read employee ID from session values.
		// Type assert to int.
		employeeID := session.Values["ID"].(int)
		err = UpdateWastageEntry(
			s.db,
			id,
			editWastageForm.WastageDate,
			editWastageForm.WastageQuantity,
			editWastageForm.WastageReason,
			editWastageForm.ProductID,
			employeeID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// GetWastagesReportMonthly handles query for monthly report
func (s *Server) GetWastagesReportMonthly(w http.ResponseWriter, r *http.Request) {
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

	reportMonthly, err := QueryWastageReportMonthly(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set header to parse response as JSON.
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(reportMonthly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
}