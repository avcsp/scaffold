package main

import (
	"fmt"
	"scaffold/internal/config"
	"scaffold/internal/router"
	"scaffold/internal/server"
)

func main() {

	// Config
	config.Load(".env")

	// Initialize
	Router := router.New()

	// Routes
	Routes(Router)

	// Dispatch
	Server := server.New(Router, fmt.Sprintf("0.0.0.0:%s", config.Getenv("PORT", "8080")))
	Server.Start()
}
