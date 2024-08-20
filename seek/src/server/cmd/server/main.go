package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/connorkuljis/seek-js/gemini"
)

func main() {
	gemApiKey := os.Getenv("GEMINIAPIKEY")
	if gemApiKey == "" {
		log.Fatal("Error. No google gemini API key! Please set env var `export G_API={your key here}`")
	}

	g, err := gemini.NewGeminiClient(gemApiKey, "gemini-1.5-flash")
	if err != nil {
		log.Fatal(err)
	}

	server := http.NewServeMux()
	server.HandleFunc("/gen", genHandler(g))
	log.Println("Starting server on 6969")
	log.Fatal(http.ListenAndServe(":6969", server))
}

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

		decoder := json.NewDecoder(r.Body)
		var msg payload
		err := decoder.Decode(&msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		g.Logger.Info("decoded message", "msg", msg)

		resumePath := "documents/resume.pdf"
		f, err := os.Open(resumePath)
		if err != nil {
			// TODO: fix error code
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer f.Close()

		// TODO: defer client.DeleteFile(*ctx, resume.Name)
		gf, err := g.UploadFile(f, nil)
		if err != nil {
			// TODO: fix error code
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		p := gemini.ResumePromptWrapper(msg.JobDescription, gf)

		resp, err := g.GenContent(p)
		if err != nil {
			// TODO: fix error code
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result := gemini.ToString(resp)

		w.Write([]byte(result))
	}
}
