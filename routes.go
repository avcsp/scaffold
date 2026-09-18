package main

import (
	"fmt"
	"net/http"

	sctx "scaffold/internal/context"
	"scaffold/internal/router"
)

func Routes(r *router.Router) {

	// Groups
	r.Group("/v1.0", func(rg *router.Group) {

		rg.Get("/users", func(c *sctx.Context) {
			c.JSON(http.StatusOK, map[string]string{"message": "list users"})
		})

		rg.Get("/users/{id}", func(c *sctx.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("user: %s", id)})
		})

	})

}
