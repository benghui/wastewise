package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs request methods & path & duration
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// SessionMiddleware is for passing session data from store to request context
func (s *Server) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Loads the session data from cookiestore.
		session, err := s.store.Get(r, "sessionCookie")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Sets the session context
		r = r.WithContext(context.WithValue(r.Context(), SessionKey{}, session))

		next.ServeHTTP(w, r)
	})
}
