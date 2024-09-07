package main

import (
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
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
	// environment variables
	EnvGeminiAPIKey string = "GEMINIAPIKEY"
	SessionAuthKey  string = "aaaaaaaaaaaaaa" // TODO: setup env var
	GotenbergURL    string = "http://127.0.0.1:3000"

	// wwwroot dirs
	StaticDir    string = "wwwroot/static"
	TemplatesDir string = "wwwroot/templates"
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

	// setup template registry
	r, err := tr.NewTemplateRegistry(wwwroot, TemplatesDir)
	if err != nil {
		log.Fatal(err)
	}
	e.Renderer = r

	// inject deps into handler
	h := handler.Handler{
		Logger:              logger,
		GeminiService:       g,
		GotenbergServiceURL: GotenbergURL,
	}

	// routes
	e.GET("/generate", h.GeneratePageGet)
	e.POST("/generate", h.GeneratePagePost)
	e.GET("/upload", h.UploadPageGet)
	e.POST("/upload", h.UploadPagePost)
	e.GET("/upload/confirm", h.ConfirmationPage)
	// e.GET("/api/pdf/:id", h.CoverLetterPDFGet)

	e.GET("/foo", func(c echo.Context) error {
		return c.Render(http.StatusOK, "partial-foo", nil)
	})
	e.GET("/bar", func(c echo.Context) error {
		return c.Render(http.StatusOK, "partial-bar", nil)
	})

	log.Fatal((e.Start(":" + port)))
}
