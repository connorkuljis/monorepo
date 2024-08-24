package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/connorkuljis/seek-js/internal/gemini"
	"github.com/gorilla/sessions"
)

// TODO: logging middleware.
// TODO: session middlware.
// TODO: parse index page from a file.
// TODO: add htmx.
// TODO: chose model option.
// TODO: what happens if URI points to a deleted file?
// TODO: check correct respone code for file upload.
func main() {
	gemApiKey := os.Getenv("GEMINIAPIKEY")
	if gemApiKey == "" {
		log.Fatal("missing environment variable [GEMINIAPIKEY]")
	}

	g, err := gemini.NewGeminiClient(gemApiKey, "gemini-1.5-flash")
	if err != nil {
		g.Logger.Error("error creating new gemini client", "message", err.Error())
		os.Exit(1)
	}

	server := http.NewServeMux()
	store := sessions.NewCookieStore([]byte("aaaaaaaaaaaaaa"))

	server.HandleFunc("/", indexHandler())
	server.HandleFunc("/gen", generateContentHandler(g, store))
	server.HandleFunc("/upload", uploadFileHandler(g, store))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		g.Logger.Info("defaulting to port", "port", port)
	}

	g.Logger.Info("started listening on port", "port", port)
	err = http.ListenAndServe(":"+port, server)
	if err != nil {
		g.Logger.Error("error listening an serving", "port", port, "message", err.Error())
		os.Exit(1)
	}
}

func indexHandler() http.HandlerFunc {
	page, err := template.New("index").Parse(`
<h1>hello</h1>
<form action="/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="pdfFile" accept=".pdf" required>
    <button type="submit">Upload</button>
</form>
<form method='post' action='/gen' enctype='application/json'>
	<input id ='text' type='text' name='description' placeholder='enter job description'/>
</form>`)
	if err != nil {
		log.Fatal(err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		page.Execute(w, nil)
	}
}

func generateContentHandler(g *gemini.GeminiClient, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			g.Logger.Warn("method_not_allowed", "expected_post")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")

		sess, err := store.Get(r, "session")
		if err != nil {
			g.Logger.Error("error_getting_session", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		uri, ok := sess.Values["uri"].(string)
		if !ok {
			err := fmt.Errorf("invalid session: no value for 'uri'")
			g.Logger.Error("invalid_session", "missing_uri", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err = r.ParseForm()
		if err != nil {
			g.Logger.Error("error_parsing_form", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		jobDescription := r.FormValue("description")
		if jobDescription == "" {
			err := fmt.Errorf("empty form field: 'description'")
			g.Logger.Error("missing_jobdescription", err)
			http.Error(w, "Bad request (missing description)", http.StatusBadRequest)
			return
		}

		p := gemini.ResumePromptWrapper(jobDescription, uri)

		resp, err := g.GenerateContent(p)
		if err != nil {
			g.Logger.Error("error_generating_content", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = sessions.Save(r, w)
		if err != nil {
			g.Logger.Error("error_saving_session", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(gemini.ToString(resp)))
	}
}

func uploadFileHandler(g *gemini.GeminiClient, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			g.Logger.Warn("method_not_allowed", "expected_post")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sess, err := store.Get(r, "session")
		if err != nil {
			g.Logger.Error("Error getting session:", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		file, header, err := r.FormFile("pdfFile")
		if err != nil {
			g.Logger.Error("Error getting file:", err)
			http.Error(w, "File error", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if filepath.Ext(header.Filename) != ".pdf" {
			g.Logger.Error("Invalid file format:", header.Filename)
			http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		gf, err := g.UploadFile(file, nil)
		if err != nil {
			g.Logger.Error("Error uploading file:", err)
			http.Error(w, "Upload error", http.StatusInternalServerError)
			return
		}

		sess.Values["uri"] = gf.URI

		err = sessions.Save(r, w)
		if err != nil {
			g.Logger.Error("Error saving session:", err)
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
