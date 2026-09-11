package main

import (
	"fmt"
	"net/http"

	"scaffold/internal/router"
)

func Routes(r *router.Router) {

	// Groups
	r.Group("/v1.0", func(rg *router.Group) {

		rg.Get("/users", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("list users"))
		})

		rg.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
			id := req.PathValue("id")
			w.Write([]byte(fmt.Sprintf("user: %s", id)))
		})

	})

}
