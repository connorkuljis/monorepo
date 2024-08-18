package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func main() {
	gemApiKey := os.Getenv("G_API")
	if gemApiKey == "" {
		log.Fatal("Error. No google gemini API key! Please set env var `export G_API={your key here}`")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(gemApiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	server := http.NewServeMux()

	server.HandleFunc("/gen", genHandler(&ctx, client))

	log.Println("Starting server on 6969")
	log.Fatal(http.ListenAndServe(":6969", server))
}

func genHandler(ctx *context.Context, client *genai.Client) http.HandlerFunc {
	type payload struct {
		JobDescription string `json:"description"`
	}

	var (
		targetModel = "gemini-1.5-flash"
		prompt      = "Please write a one-page cover letter for the job description and resume."
		resumePath  = "documents/resume.pdf"
	)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

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
		logger.Info("decoded message", "msg", msg)

		// remember to delete the file from the client after
		resume, err := uploadDocument(*ctx, client, resumePath, nil)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info("uploaded file", "resume", resume)

		m := client.GenerativeModel(targetModel)
		logger.Info("selected model:", "name", targetModel)

		combinedPrompt := []genai.Part{
			genai.Text(prompt),
			genai.Text(msg.JobDescription),
			genai.FileData{URI: resume.URI},
		}
		logger.Info("combined prompt", "prompt", combinedPrompt)

		// Generate content using the prompt.
		mResp, err := m.GenerateContent(*ctx, combinedPrompt...)
		if err != nil {
			log.Fatal(err)
		}
		logger.Info("generated content", "generatedContentResponse", mResp)

		// Handle the response of generated text
		var results []string
		for _, c := range mResp.Candidates {
			if c.Content != nil {
				var result string
				for _, part := range c.Content.Parts {
					result += fmt.Sprint(part)
				}
				results = append(results, result)
			}
		}
		logger.Info("result", "result", results)

		err = client.DeleteFile(*ctx, resume.Name)
		if err != nil {
			logger.Error("Failed to delete file", "file", resume.Name, "error", err.Error())
			http.Error(w, "Unable to process your request due to an internal error", http.StatusInternalServerError)
			return
		}

		_, err = w.Write([]byte(results[0]))
		if err != nil {
			logger.Error("Unable to write data to connection", "w", w, "error", err.Error())
			return
		}

		return
	}
}

// Upload a document useing the file api
func uploadDocument(ctx context.Context, client *genai.Client, filename string, opts *genai.UploadFileOptions) (*genai.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := client.UploadFile(ctx, "", f, opts)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
