package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// LoginEmployee authenticates an employee
func (s *Server) LoginEmployee(w http.ResponseWriter, r *http.Request) {
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

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

	} else {
		http.Error(w, "Invalid content-type", http.StatusBadRequest)
	}
}

// LogoutEmployee revokes authentication.
func (s *Server) LogoutEmployee(w http.ResponseWriter, r *http.Request) {
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	session.Values["user"] = nil
	session.Values["ID"] = nil
	session.Values["auth"] = nil
	session.Values["role"] = nil
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateEmployee handles creating an employee
func (s *Server) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	// Check that employee is authenticated.
	if session.Values["auth"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check that user role is authorized as admin
	if session.Values["role"].(string) != "admin" {
		http.Error(w, "No Access", http.StatusForbidden)
		return
	}

	if r.Header.Get("Content-type") == "application/json" {
		newEmployee := &Employees{}

		// Locks & defer Unlock to prevent race conditions
		newEmployee.Mu.Lock()
		defer newEmployee.Mu.Unlock()

		err := json.NewDecoder(r.Body).Decode(newEmployee)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hashValue, err := strconv.Atoi(os.Getenv("HASH_VALUE"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Hash password
		bPassword, err := bcrypt.GenerateFromPassword([]byte(newEmployee.Password), hashValue)
		if err != nil {
			log.Panic(err)
		}
		// Panic recovery
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		err = CreateEmployeeEntry(
			s.db,
			newEmployee.Username,
			newEmployee.Firstname,
			newEmployee.Lastname,
			string(bPassword),
			newEmployee.Role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	} else {
		http.Error(w, "Invalid content-type", http.StatusBadRequest)
	}
}

// GetProducts handles query to all products data.
// It is a protected resource requiring authentication value from sessions.
func (s *Server) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	// Check that employee is authenticated.
	if session.Values["auth"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	// Check that employee is authenticated.
	if session.Values["auth"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	// Check that employee is authenticated.
	if session.Values["auth"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Invalid content-type", http.StatusBadRequest)
	}
}

// GetSingleWastage handles query for wastage entry.
func (s *Server) GetSingleWastage(w http.ResponseWriter, r *http.Request) {
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	// Check that employee is authenticated.
	if session.Values["auth"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
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
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	// Check that employee is authenticated.
	if session.Values["auth"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Invalid content-type", http.StatusBadRequest)
	}
}

// GetWastagesReportMonthly handles query for monthly report
func (s *Server) GetWastagesReportMonthly(w http.ResponseWriter, r *http.Request) {
	// Retrieve session from request context.
	session := r.Context().Value(SessionKey{}).(*sessions.Session)

	// Check that employee is authenticated.
	if session.Values["auth"] != true {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
