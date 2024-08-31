package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/google/generative-ai-go/genai"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (h *Handler) UploadResumePost(c echo.Context) error {
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
	gf, err := h.GeminiService.UploadFile(f, "", opts)
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
