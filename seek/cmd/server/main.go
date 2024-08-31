package main

import (
	"html/template"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/connorkuljis/seek-js/cmd/server/handlers"
	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TODO: create and handle env var for the pdf generation url, also need to deploy to gcr.
// TODO: name the pdf file with a formatted name, need db?
// TODO: include: contact and email in cv.

const (
	EnvGeminiAPIKey string = "GEMINIAPIKEY"
	GeminiModelName string = "gemini-1.5-flash"
	SessionAuthKey  string = "aaaaaaaaaaaaaa"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

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

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(SessionAuthKey))))

	e.Renderer = &Template{template.Must(template.ParseGlob("templates/*.html"))}

	h := handlers.Handler{
		Logger:        logger,
		GeminiService: g,
	}

	e.GET("/", h.IndexHandlerGet)
	e.GET("/upload", h.UploadResumeGet)
	e.POST("/upload", h.UploadResumePost)
	e.POST("/api/gen", h.GenerateCoverLetterPost)
	e.GET("/api/pdf/:id", h.CoverLetterPDFGet)

	err = e.Start(":" + port)
	if err != nil {
		logger.Error("error listening an serving", "port", port, "message", err.Error())
		os.Exit(1)
	}
}
