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

func (form *Form) Validate() error {
	if form.description == "" || form.place == "" || form.email == "" || form.phone == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing form value")
	}

	if form.targetModel != "gemini-1.5-flash" && form.targetModel != "gemini-1.5-pro" {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported model")
	}

	return nil
}

func (s *Server) GeneratePagePost(c echo.Context) error {
	// 1. Validating user-session and form values
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

	// b, err := newCoverLetter.Render(c)
	// if err != nil {
	// 	return err
	// }

	// 	// 4.  Opening the rendered cover letter to send to gotenburg to format as pdf
	// 	// NOTE: this is a hack as goteberg does not allow a converting html bytes directly to pdf as it must be a form file in the header.
	// 	filename = filepath.Join("out", filename, "index.html")
	// 	file, err := os.Open(filename)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer file.Close()

	// 	body := &bytes.Buffer{}
	// 	writer := multipart.NewWriter(body)

	// 	// 4.1 Allocating and copying the form fil
	// 	part, err := writer.CreateFormFile("files", filepath.Base(filename))
	// 	if err != nil {
	// 		return err
	// 	}

	// 	_, err = io.Copy(part, file)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = writer.Close()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// 4.2 Posting the form file to gotenberg
	// 	url := h.GotenbergServiceURL + "/forms/chromium/convert/html"
	// 	req, err := http.NewRequest("POST", url, body)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 	client := &http.Client{}
	// 	resp2, err := client.Do(req)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer resp2.Body.Close()
	// 	if resp2.StatusCode != http.StatusOK {
	// 		return err
	// 	}

	// 	// 4.3 Respond with pdf
	// 	c.Response().Header().Set("Content-Type", "application/pdf")
	// 	c.Response().Header().Set("Content-Disposition", "attachment; filename=converted.pdf")

	// 	_, err = io.Copy(c.Response(), resp2.Body)
	// 	if err != nil {
	// 		return err
	// 	}

	return c.Render(http.StatusOK, "cover-letter.html", map[string]any{"CoverLetter": newCoverLetter})
}
