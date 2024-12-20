package server

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/connorkuljis/seek-js/internal/store"
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

type Server struct {
	Env                 *Env
	Echo                *echo.Echo
	DB                  *store.DB
	Logger              *slog.Logger
	GeminiClient        *gemini.GeminiClient
	GotenbergServiceURL string
	WWWRoot             embed.FS
}

type Env struct {
	Port         string
	GeminiAPIKey string
}

func NewServer(env *Env, wwwroot embed.FS, db *store.DB, logger *slog.Logger, geminiClient *gemini.GeminiClient) (*Server, error) {
	server := &Server{
		Env:                 env,
		Echo:                echo.New(),
		DB:                  db,
		Logger:              logger,
		GeminiClient:        geminiClient,
		GotenbergServiceURL: GotenbergURL,
		WWWRoot:             wwwroot,
	}

	// Use our custom renderer / registered templates
	renderer, err := tr.NewTemplateRegistry(wwwroot, TemplatesDir)
	if err != nil {
		return nil, err
	}
	server.Echo.Renderer = renderer

	return server, nil
}

func (s *Server) Routes() error {
	staticFS, err := fs.Sub(s.WWWRoot, StaticDir)
	if err != nil {
		return err
	}

	e := s.Echo
	e.StaticFS("/", staticFS)
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/login")
	})

	e.GET("/login", s.LoginHandler)
	e.POST("/login", s.LoginPostHandler)

	e.GET("/account", s.AccountHandler)

	e.GET("/generate", s.GeneratePageGet)
	e.POST("/generate", s.GeneratePagePost)
	e.GET("/upload", s.UploadPageGet)
	e.POST("/upload", s.UploadPagePost)

	e.GET("/upload/confirm", s.ConfirmationPage)
	e.GET("/api/pdf/:id", s.CoverLetterPDFGet)

	return nil
}

func (s *Server) Middleware() {
	e := s.Echo
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	store := sessions.NewCookieStore([]byte(SessionAuthKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,
		Secure:   false,
		Domain:   "",
	}
	e.Use(session.Middleware(store))
}

func (s *Server) Start() error {
	s.Logger.Info("Starting Server")
	return s.Echo.Start("0.0.0.0:" + s.Env.Port)
}

func LoadEnvVars() (*Env, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	geminiAPIKey := os.Getenv(EnvGeminiAPIKey)
	if geminiAPIKey == "" {
		return nil, fmt.Errorf("Missing $GEMINIAPIKEY")
	}

	return &Env{
		Port:         port,
		GeminiAPIKey: geminiAPIKey,
	}, nil
}
