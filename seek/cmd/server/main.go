package main

import (
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/google/generative-ai-go/genai"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TODO: check if uri expired?
// TODO: include: contact and email in cv.
// TODO: stream results back with websockets

const (
	EnvGeminiAPIKey string = "GEMINIAPIKEY"
	GeminiModelName string = "gemini-1.5-flash"
	SessionAuthKey  string = "aaaaaaaaaaaaaa"
)

type Handler struct {
	G *gemini.GeminiClient
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

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

	store := sessions.NewCookieStore([]byte(SessionAuthKey))

	h := &Handler{g}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(store))

	e.Renderer = &Template{template.Must(template.ParseGlob("templates/*.html"))}

	e.GET("/", h.IndexHandler)
	e.POST("/gen", h.GenerateContentHandler)
	e.POST("/upload", h.UploadFileHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		g.Logger.Info("defaulting to port", "port", port)
	}

	err = e.Start(":" + port)
	if err != nil {
		g.Logger.Error("error listening an serving", "port", port, "message", err.Error())
		os.Exit(1)
	}
}

func (h *Handler) IndexHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	uri, ok := sess.Values["uri"].(string)
	if !ok {
		// return echo.NewHTTPError(http.StatusUnauthorized, "Please provide a uri")
	}
	filename, ok := sess.Values["filename"].(string)
	if !ok {
		// TODO:
	}

	return c.Render(http.StatusOK, "index", map[string]string{
		"URI":      uri,
		"Filename": filename,
	})
}

func (h *Handler) GenerateContentHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	uri, ok := sess.Values["uri"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please provide a uri")
	}

	jobDescription := c.FormValue("description")
	if jobDescription == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty form value for: `description`")
	}

	model := c.FormValue("model")
	if err != nil {
		return err
	}

	var targetModel gemini.Model
	switch model {
	case "gemini-1.5-flash":
		targetModel = gemini.Flash
	case "gemini-1.5-pro":
		targetModel = gemini.Pro
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported model")
	}

	p := gemini.ResumePromptWrapper(jobDescription, uri)

	resp, err := h.G.GenerateContent(p, targetModel)
	if err != nil {
		return err
	}

	cv, err := gemini.ParseCoverLetterJSON(gemini.ToString(resp))
	if err != nil {
		return err
	}

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "cover-letter", cv)
}

func (h *Handler) UploadFileHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	fileHeader, err := c.FormFile("pdfFile")
	if err != nil {
		return err
	}

	if filepath.Ext(fileHeader.Filename) != ".pdf" {
		return echo.NewHTTPError(http.StatusBadRequest, "only .pdf files are permited")
	}

	f, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	opts := &genai.UploadFileOptions{DisplayName: fileHeader.Filename}
	gf, err := h.G.UploadFile(f, "", opts)
	if err != nil {
		return err
	}

	sess.Values["uri"] = gf.URI
	sess.Values["filename"] = fileHeader.Filename

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}
