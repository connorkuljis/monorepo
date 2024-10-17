package server

import (
	"net/http"

	"github.com/connorkuljis/seek-js/internal/store"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func (s *Server) LoginHandler(c echo.Context) error {

	return c.Render(http.StatusOK, "login.html", nil)
}

func (s *Server) LoginPostHandler(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	email := c.FormValue("email")

	user := &store.User{
		Email: email,
	}

	repo := store.NewUserRepository(s.DB.Connection())

	err = repo.Create(user)
	if err != nil {
		return err
	}

	s.Logger.Info("Created user", "email", user.Email, "id", user.ID)

	sess.Values["id"] = user.ID

	err = sess.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}
