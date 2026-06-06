package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler)

	log.Fatal(http.ListenAndServe(":8080", r))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}
