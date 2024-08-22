package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/connorkuljis/seek-js/gemini"
)

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
	server.HandleFunc("/", indexHandler())
	server.HandleFunc("/gen", genHandler(g))

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

// TODO: support both form data and json data
// TODO: implement caching resume data by using existing URI
// TODO: flag for pro or flash model selection
func genHandler(g *gemini.GeminiClient) http.HandlerFunc {
	type payload struct {
		JobDescription string `json:"description"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		r.ParseForm()

		var msg payload
		msg.JobDescription = r.FormValue("description")
		g.Logger.Info("got description", "description", msg.JobDescription)

		// decoder := json.NewDecoder(r.Body)
		// var msg payload
		// err := decoder.Decode(&msg)
		// if err != nil {
		// 	e := fmt.Errorf("error decoding request body: %w", err)
		// 	g.Logger.Error("bad request", "error", e.Error(), "body", r.Body)
		// 	http.Error(w, err.Error(), http.StatusBadRequest)
		// 	return
		// }
		// g.Logger.Info("decoded message", "msg", msg)

		// Below this line is duplicated in cli, but returning errors differently

		f, err := os.Open("static/Connor-Kuljis_Resume_2024-07.pdf")
		if err != nil {
			e := fmt.Errorf("internal server error: %w", err)
			g.Logger.Error("error", "error", e.Error())
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		gf, err := g.UploadFile(f, nil)
		if err != nil {
			e := fmt.Errorf("internal server error: %w", err)
			g.Logger.Error("error", "error", e.Error())
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
		defer g.Client.DeleteFile(*g.Ctx, gf.Name)

		p := gemini.ResumePromptWrapper(msg.JobDescription, gf)

		resp, err := g.GenerateContent(p)
		if err != nil {
			e := fmt.Errorf("internal server error: %w", err)
			http.Error(w, e.Error(), http.StatusInternalServerError)
			g.Logger.Error("error", "error", e.Error())
			return
		}

		w.Write([]byte(gemini.ToString(resp)))
	}
}

func indexHandler() http.HandlerFunc {
	page, err := template.New("index").Parse(`
		<h1>hello</h1>
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
