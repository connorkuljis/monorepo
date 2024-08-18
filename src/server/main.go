package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Set header to allow POST or OPTION
		w.Header().Set("Access-Control-Allow-Origin", "*")

		decoder := json.NewDecoder(r.Body)
		var msg payload
		err := decoder.Decode(&msg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("Msg:", msg)

		resumePath := "documents/resume.pdf"
		resume, err := uploadDocument(*ctx, client, resumePath, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer client.DeleteFile(*ctx, resume.Name)

		log.Println("Uploaded:", resume)

		model := client.GenerativeModel("gemini-1.5-flash")

		combinedPrompt := []genai.Part{
			genai.Text("Please write a one-page cover letter for the job description and resume."),
			genai.Text(msg.JobDescription),
			genai.FileData{URI: resume.URI},
		}
		log.Println("Prompt:", combinedPrompt)

		// Generate content using the prompt.
		mResp, err := model.GenerateContent(*ctx, combinedPrompt...)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Response:", mResp)

		// Handle the response of generated text
		for _, c := range mResp.Candidates {
			if c.Content != nil {
				fmt.Fprintln(w, *c.Content)
			}
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
