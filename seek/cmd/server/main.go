package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/connorkuljis/seek-js/internal/cv"
	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/google/generative-ai-go/genai"
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

type Handler struct {
	GClient *gemini.GeminiClient
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
	e.GET("/upload", h.UploadHandler)
	e.POST("/upload", h.UploadFileHandler)
	e.POST("/api/gen", h.GenerateContentHandler)
	e.GET("/api/pdf/:id", h.GeneratePDF)

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
		c.Redirect(http.StatusSeeOther, "/upload")
	}

	filename, ok := sess.Values["filename"].(string)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/upload")
	}

	return c.Render(http.StatusOK, "index", map[string]string{
		"URI":      uri,
		"Filename": filename,
	})
}

func (h *Handler) UploadHandler(c echo.Context) error {
	_, err := session.Get("session", c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "upload", nil)
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
	gf, err := h.GClient.UploadFile(f, "", opts)
	if err != nil {
		return err
	}

	sess.Values["uri"] = gf.URI
	sess.Values["filename"] = gf.Name

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
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

	filename, ok := sess.Values["filename"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please provide a filename")
	}

	jobDescription := c.FormValue("description")
	if jobDescription == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty form value for: `description`")
	}

	model := c.FormValue("model")

	var targetModel gemini.Model
	switch model {
	case "gemini-1.5-flash":
		targetModel = gemini.Flash
	case "gemini-1.5-pro":
		targetModel = gemini.Pro
	case "":
		return echo.NewHTTPError(http.StatusBadRequest, "missing model")
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported model")
	}

	p := gemini.ResumePromptWrapper(cv.Prompt, jobDescription, uri)

	resp, err := h.GClient.GenerateContent(p, targetModel)
	if err != nil {
		return err
	}

	coverLetter, err := cv.NewCoverLetterFromJSON(filename, gemini.ToString(resp))
	if err != nil {
		return err
	}

	err = coverLetter.SaveAsHTML()
	if err != nil {
		return err
	}

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "cover-letter", coverLetter)
}

func (h *Handler) GeneratePDF(c echo.Context) error {
	_, err := session.Get("session", c)
	if err != nil {
		return err
	}

	id := c.Param("id") // this is not safe btw

	filename := filepath.Join("out", id, "index.html")

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("files", filepath.Base(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	url := "http://127.0.0.1:3000/forms/chromium/convert/html"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err
	}

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=converted.pdf")

	_, err = io.Copy(c.Response(), resp.Body)
	if err != nil {
		return err
	}

	return nil
}
