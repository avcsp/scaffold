package main

import (
	"fmt"
	"net/http"

	"scaffold/internal/engine"
	"scaffold/internal/router"
)

func Routes(r *router.Router) {

	// Groups
	r.Group("/v1.0", func(r *router.Group) {

		r.Get("/users", func(c *engine.Context) {
			c.JSON(http.StatusOK, map[string]string{"message": "list users"})
		})

		r.Get("/users/{id}", func(c *engine.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("user: %s", id)})
		})

	})

}
