package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/connorkuljis/seek-js/internal/db"
	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/connorkuljis/seek-js/internal/server"
)

//go:embed wwwroot
var wwwroot embed.FS

func main() {
	start()
}

func start() {
	// Logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Environment variables
	env, err := server.LoadEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	// Add gemini client for gen ai
	ctx := context.Background()
	gc, err := gemini.NewClient(&ctx, env.GeminiAPIKey, logger)
	if err != nil {
		log.Fatal(err)
	}

	db, err := db.NewDB(db.Config{Name: "application.db", Directory: "db"})
	defer db.Close()

	err = db.InitSchema()
	if err != nil {
		log.Fatal(err)
	}

	// Create and start the server
	server, err := server.NewServer(env, wwwroot, db, logger, gc)
	if err != nil {
		log.Fatal(err)
	}
	server.Middleware()
	server.Routes()
	log.Fatal(server.Start())
}
