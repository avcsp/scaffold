package main

import (
	"fmt"
	"log"

	"scaffold/internal/config"
	"scaffold/internal/database"
	"scaffold/internal/logger"
	"scaffold/internal/router"
	"scaffold/internal/server"
)

func main() {

	// Config
	config.Load(".env")

	// Logger
	logs := logger.Init()

	// Initialize
	Router := router.New()

	// Database
	db, err := database.Connect(logs.Scaffold)
	if err != nil {
		log.Fatal(err)
	}

	// Routes
	Routes(Router)

	// Dispatch
	Server := server.New(Router, fmt.Sprintf("0.0.0.0:%s", config.Getenv("PORT", "8080")), logs.Scaffold)

	// Hooks
	Server.OnShutdown(func() { database.Close(db) })

	// Start
	Server.Start()
}
