package main

import (
	"scaffold/internal/router"
	"scaffold/internal/server"
)

func main() {

	// Initialize
	Router := router.New()

	// Routes
	Routes(Router)

	// Dispatch
	Server := server.New(Router, ":8080")
	Server.Start()
}
