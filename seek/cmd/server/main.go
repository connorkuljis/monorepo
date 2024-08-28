package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TODO: [ui] return results using htmx OR stream them in.
// TODO: [ui] chose model option.
// TODO: [ui] styling.
// TODO: what happens if URI points to a deleted file? -> internal server error: do not have perms or deleted.

const (
	EnvGeminiAPIKey string = "GEMINIAPIKEY"
	GeminiModelName string = "gemini-1.5-flash"
	SessionAuthKey  string = "aaaaaaaaaaaaaa"
)

type Handler struct {
	G *gemini.GeminiClient
}

func main() {
	geminiAPIKey := os.Getenv(EnvGeminiAPIKey)
	if geminiAPIKey == "" {
		log.Fatalf("Error: Required environment variable %s is not set.\n"+
			"This variable is necessary to connect to the gemini api.\n"+
			"Please set %s before running the application.\n"+
			"Example: export %s=<value>", EnvGeminiAPIKey, EnvGeminiAPIKey, EnvGeminiAPIKey)
	}

	g, err := gemini.NewGeminiClient(geminiAPIKey, GeminiModelName)
	if err != nil {
		g.Logger.Error("error creating new gemini client", "message", err.Error())
		os.Exit(1)
	}

	store := sessions.NewCookieStore([]byte(SessionAuthKey))

	// server.HandleFunc("/gen", generateContentHandler(g, store))
	// server.HandleFunc("/upload", uploadFileHandler(g, store))

	h := &Handler{g}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(store))

	e.GET("/", h.IndexHandler)
	e.POST("/gen", h.GenerateContentHandler)
	e.POST("/upload", h.UploadFileHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		g.Logger.Info("defaulting to port", "port", port)
	}

	// g.Logger.Info("started listening on port", "port", port)
	err = e.Start(":" + port)
	if err != nil {
		g.Logger.Error("error listening an serving", "port", port, "message", err.Error())
		os.Exit(1)
	}
}

func (h *Handler) IndexHandler(c echo.Context) error {
	t, err := template.New("index").ParseFiles("templates/index.html")
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "index", nil)
	if err != nil {
		return err
	}

	return c.HTML(http.StatusOK, tpl.String())
}

func (h *Handler) GenerateContentHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		// g.Logger.Error("error_getting_session", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		// return
		return err
	}

	uri, ok := sess.Values["uri"].(string)
	if !ok {
		// err := fmt.Errorf("invalid session: no value for 'uri'")
		// g.Logger.Error("invalid_session", "missing_uri", err)
		// http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return echo.NewHTTPError(http.StatusUnauthorized, "Please provide a uri")
	}

	// err = r.ParseForm()
	// if err != nil {
	// 	g.Logger.Error("error_parsing_form", err)
	// 	http.Error(w, "Bad request", http.StatusBadRequest)
	// 	return
	// }

	jobDescription := c.FormValue("description")
	if jobDescription == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "empty form value for: `description`")
	}

	p := gemini.ResumePromptWrapper(jobDescription, uri)

	resp, err := h.G.GenerateContent(p)
	if err != nil {
		return err
	}

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		// g.Logger.Error("error_saving_session", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		return err
	}

	return c.HTML(http.StatusOK, gemini.ToString(resp))
}

func (h *Handler) UploadFileHandler(c echo.Context) error {
	// if r.Method != http.MethodPost {
	// 	g.Logger.Warn("method_not_allowed", "expected_post")
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }

	sess, err := session.Get("session", c)
	if err != nil {
		// g.Logger.Error("Error getting session:", err)
		// http.Error(w, "Session error", http.StatusInternalServerError)
		return err
	}

	// fileHeader, fileHeader, err := r.FormFile("pdfFile")
	fileHeader, err := c.FormFile("pdfFile")
	if err != nil {
		// g.Logger.Error("Error getting file:", err)
		// http.Error(w, "File error", http.StatusBadRequest)
		return err
	}
	// defer file.Close()

	if filepath.Ext(fileHeader.Filename) != ".pdf" {
		// g.Logger.Error("Invalid file format:", header.Filename)
		// http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return echo.NewHTTPError(http.StatusBadRequest, "only .pdf files are permited")
	}

	f, err := fileHeader.Open()
	if err != nil {
		return err
	}

	gf, err := h.G.UploadFile(f, nil)
	if err != nil {
		// g.Logger.Error("Error uploading file:", err)
		// http.Error(w, "Upload error", http.StatusInternalServerError)
		return err
	}

	sess.Values["uri"] = gf.URI

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		// g.Logger.Error("Error saving session:", err)
		// http.Error(w, "Session error", http.StatusInternalServerError)
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}
