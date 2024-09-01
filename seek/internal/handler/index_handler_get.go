package handler

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (h *Handler) IndexHandlerGet(c echo.Context) error {
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
