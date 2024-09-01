package main

import (
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"os"

	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/connorkuljis/seek-js/internal/handler"
	tr "github.com/connorkuljis/seek-js/internal/template_registry"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	EnvGeminiAPIKey string = "GEMINIAPIKEY"
	SessionAuthKey  string = "aaaaaaaaaaaaaa"

	StaticDir    = "wwwroot/static"
	TemplatesDir = "wwwroot/templates"
)

//go:embed wwwroot
var wwwroot embed.FS

// TODO: create and handle env var for the pdf generation url, also need to deploy to gcr.
// TODO: name the pdf file with a formatted name, need db?
// TODO: include: contact and email in cv.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		logger.Info("defaulting to port", "port", port)
	}

	geminiAPIKey := os.Getenv(EnvGeminiAPIKey)
	if geminiAPIKey == "" {
		log.Fatalf("Error: Required environment variable %s is not set.\n"+
			"This variable is necessary to connect to the gemini api.\n"+
			"Please set %s before running the application.\n"+
			"Example: export %s=<value>", EnvGeminiAPIKey, EnvGeminiAPIKey, EnvGeminiAPIKey)
	}

	g, err := gemini.NewGeminiClient(geminiAPIKey, logger)
	if err != nil {
		g.Logger.Error("error creating new gemini client", "message", err.Error())
		os.Exit(1)
	}

	e := echo.New()

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SessionAuthKey))))

	// static content
	static, err := fs.Sub(wwwroot, StaticDir)
	if err != nil {
		log.Fatal(err)
	}
	e.StaticFS("/", static)

	// template registry
	r, err := tr.NewTemplateRegistry(wwwroot, TemplatesDir)
	if err != nil {
		log.Fatal(err)
	}
	e.Renderer = r

	// handler
	h := handler.Handler{
		Logger:        logger,
		GeminiService: g,
	}

	// routes
	e.GET("/", h.IndexHandlerGet)
	e.GET("/upload", h.UploadResumeGet)
	e.POST("/upload", h.UploadResumePost)
	e.POST("/api/gen", h.GenerateCoverLetterPost)
	e.GET("/api/pdf/:id", h.CoverLetterPDFGet)

	log.Fatal((e.Start(":" + port)))
}
