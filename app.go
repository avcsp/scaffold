package main

import (
	"fmt"
	"net/http"

	"scaffold/internal/router"
	"scaffold/internal/server"
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

	// Dispatch
	Server := server.New(Router, ":8080")
	Server.Start()
}
