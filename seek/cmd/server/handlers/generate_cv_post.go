package handlers

import (
	"bytes"
	"net/http"

	"github.com/connorkuljis/seek-js/internal/cv"
	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GenerateCoverLetterPost(c echo.Context) error {
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

	resp, err := h.GeminiService.GenerateContent(p, targetModel)
	if err != nil {
		return err
	}

	coverLetter, err := cv.NewCoverLetterFromJSON(filename, gemini.ToString(resp))
	if err != nil {
		return err
	}

	data := map[string]any{
		"CoverLetter": coverLetter,
		"Print":       true,
	}

	var buf bytes.Buffer
	c.Echo().Renderer.Render(&buf, "cover-letter", data, c)
	err = coverLetter.SaveAsHTML(buf.Bytes())
	if err != nil {
		return err
	}

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	data["Print"] = false
	return c.Render(http.StatusOK, "cover-letter", data)
}
