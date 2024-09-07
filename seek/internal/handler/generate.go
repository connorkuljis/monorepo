package handler

import (
	"bytes"
	"net/http"

	"github.com/connorkuljis/seek-js/internal/cv"
	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/google/generative-ai-go/genai"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/iterator"
)

func (h *Handler) GeneratePageGet(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	uri, ok1 := sess.Values["uri"].(string)
	filename, ok2 := sess.Values["filename"].(string)

	if !ok1 || !ok2 {
		c.Redirect(http.StatusSeeOther, "/upload")
	}

	data := map[string]any{
		"URI":      uri,
		"Filename": filename,
	}

	return c.Render(http.StatusOK, "index", data)
}

func (h *Handler) GeneratePagePost(c echo.Context) error {
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

	// form values
	var (
		place       = c.FormValue("place")
		email       = c.FormValue("email")
		phone       = c.FormValue("phone")
		description = c.FormValue("description")
		targetModel = c.FormValue("model")
	)

	// if description == "" || place == "" || email == "" || phone == "" {
	// 	return echo.NewHTTPError(http.StatusBadRequest, "Missing form value")
	// }

	if targetModel != "gemini-1.5-flash" && targetModel != "gemini-1.5-pro" {
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported model")
	}

	parts := []genai.Part{
		genai.Text(cv.Prompt),
		genai.Text("Place where I found the job description: " + place),
		genai.Text(description),
		genai.FileData{URI: uri},
	}

	m := h.GeminiService.Client.GenerativeModel(targetModel)

	m.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	ooom := ""
	iter := m.GenerateContentStream(*h.GeminiService.Ctx, parts...)
	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		ooom += gemini.ToString(resp)
		c.Response().Writer.Write([]byte(gemini.ToString(resp)))
		c.Response().Flush()
	}

	coverLetter, err := cv.NewCoverLetterFromJSON(filename, string(ooom))
	if err != nil {
		return err
	}
	coverLetter.Email = email
	coverLetter.Phone = phone

	data := map[string]any{
		"CoverLetter": coverLetter,
	}

	// render and save a seperate html file for printing as pdf
	var buf bytes.Buffer
	if err := c.Echo().Renderer.Render(&buf, "cover-letter-print", data, c); err != nil {
		return err
	}

	if err := coverLetter.SaveAsHTML(buf.Bytes()); err != nil {
		return err
	}

	c.Response().Write(buf.Bytes())

	err = sessions.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	// // generating pdf directly
	// filename = filepath.Join("out", filename, "index.html")

	// file, err := os.Open(filename)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// body := &bytes.Buffer{}
	// writer := multipart.NewWriter(body)

	// part, err := writer.CreateFormFile("files", filepath.Base(filename))
	// if err != nil {
	// 	return err
	// }

	// _, err = io.Copy(part, file)
	// if err != nil {
	// 	return err
	// }

	// err = writer.Close()
	// if err != nil {
	// 	return err
	// }

	// url := h.GotenbergServiceURL + "/forms/chromium/convert/html"
	// req, err := http.NewRequest("POST", url, body)
	// if err != nil {
	// 	return err
	// }

	// req.Header.Set("Content-Type", writer.FormDataContentType())

	// client := &http.Client{}
	// resp2, err := client.Do(req)
	// if err != nil {
	// 	return err
	// }
	// defer resp2.Body.Close()
	// if resp2.StatusCode != http.StatusOK {
	// 	return err
	// }

	// c.Response().Header().Set("Content-Type", "application/pdf")
	// c.Response().Header().Set("Content-Disposition", "attachment; filename=converted.pdf")

	// _, err = io.Copy(c.Response(), resp2.Body)
	// if err != nil {
	// 	return err
	// }

	// return c.Redirect(http.StatusOK, "/")

	return nil
}
