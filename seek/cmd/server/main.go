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

// TODO: add structured logging to http errors.
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
	server.HandleFunc("/gen", genHandler(g, store))
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

func genHandler(g *gemini.GeminiClient, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")

		sess, err := store.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var uri string
		switch v := sess.Values["uri"].(type) {
		case string:
			uri = v
		default:
			http.Error(w, "Cannot find URI value in session. Did you upload a file?", http.StatusNotAcceptable)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jobDescription := r.FormValue("description")
		if jobDescription == "" {
			http.Error(w, "Description cannot be emtpy", http.StatusBadRequest)
			return
		}

		p := gemini.ResumePromptWrapper(jobDescription, uri)

		resp, err := g.GenerateContent(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessions.Save(r, w)

		w.Write([]byte(gemini.ToString(resp)))
	}
}

func uploadFileHandler(g *gemini.GeminiClient, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sess, err := store.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file, header, err := r.FormFile("pdfFile")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		if filepath.Ext(header.Filename) != ".pdf" {
			http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
			return
		}

		gf, err := g.UploadFile(file, nil)
		if err != nil {
			e := fmt.Errorf("internal server error: %w", err)
			g.Logger.Error("error", "error", e.Error())
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		sess.Values["uri"] = gf.URI

		err = sessions.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
