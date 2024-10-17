package server

import (
	"fmt"
	"net/http"

	"github.com/connorkuljis/seek-js/internal/store"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (s *Server) AccountHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	id, ok := sess.Values["id"].(int64)
	if !ok {
		return fmt.Errorf("missing session id")
	}

	repo := store.NewUserRepository(s.DB.Connection())

	user, err := repo.GetByID(id)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "account.html", map[string]any{"User": user})
}
