package main

import (
	"fmt"
	"log"
	"net/http"

	"scaffold/internal/router"
)

func main() {

	// Initialize
	Router := router.New()

	// Groups
	Router.Group("/v1.0", func(rg *router.Group) {

		rg.Get("/users", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("list users"))
		})

		rg.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
			id := req.PathValue("id")
			w.Write([]byte(fmt.Sprintf("user: %s", id)))
		})

	})

	// Health
	Router.Get("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("ok"))
	})

	// Dispatch
	log.Fatal(Router.Serve(":8080"))
}
