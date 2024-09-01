package handler

import (
	"log/slog"

	"github.com/connorkuljis/seek-js/internal/gemini"
)

type Handler struct {
	Logger              *slog.Logger
	GeminiService       *gemini.GeminiClient
	GotenbergServiceURL string
}
