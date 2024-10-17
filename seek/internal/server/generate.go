package server

import (
	"net/http"
	"time"

	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (h *Server) GeneratePageGet(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	_, ok1 := sess.Values["uri"].(string)
	_, ok2 := sess.Values["filename"].(string)

	if !ok1 || !ok2 {
		c.Redirect(http.StatusSeeOther, "/upload")
	}

	return c.Render(http.StatusOK, "index.html", nil)
}

type Form struct {
	place       string
	email       string
	phone       string
	description string
	targetModel string
}

func (s *Server) GeneratePagePost(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	uri, ok := sess.Values["uri"].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please provide a uri")
	}

	form := Form{
		place:       c.FormValue("place"),
		email:       c.FormValue("email"),
		phone:       c.FormValue("phone"),
		description: c.FormValue("description"),
		targetModel: c.FormValue("model"),
	}

	err = form.Validate()
	if err != nil {
		return err
	}

	newCoverLetter := gemini.NewCoverLetter(uri, "index.html", form.email, form.phone, form.description, form.place, time.Now())
	err = newCoverLetter.Generate(s.GeminiClient, form.targetModel)
	if err != nil {
		return err
	}

	// b, err := newCoverLetter.RenderHTML(c)
	// if err != nil {
	// 	return err
	// }

	return c.Render(http.StatusOK, "cover-letter.html", map[string]any{"CoverLetter": newCoverLetter})
}

func (form *Form) Validate() error {
	if form.description == "" || form.place == "" || form.email == "" || form.phone == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing form value")
	}

	if form.targetModel != "gemini-1.5-flash" && form.targetModel != "gemini-1.5-pro" {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported model")
	}

	return nil
}
