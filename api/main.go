package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	// read environment variable from .env file
	port := os.Getenv("PORT")

	router := mux.NewRouter()

	router.HandleFunc("/", helloWorld)

	listenAt := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(listenAt, router))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}
