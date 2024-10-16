package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"

	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/connorkuljis/seek-js/internal/server"
)

//go:embed wwwroot
var wwwroot embed.FS

// TODO: create and handle env var for the pdf generation url, also need to deploy to gcr.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	env, err := server.LoadEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	gc, err := gemini.NewClient(&ctx, env.GeminiAPIKey, logger)
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.NewServer(env, wwwroot, logger, gc)
	if err != nil {
		log.Fatal(err)
	}

	server.Middleware()
	server.Routes()
	server.Start()
}
