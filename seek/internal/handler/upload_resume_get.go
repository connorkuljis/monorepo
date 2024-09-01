package handler

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (h *Handler) UploadResumeGet(c echo.Context) error {
	_, err := session.Get("session", c)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "upload", nil)
}
