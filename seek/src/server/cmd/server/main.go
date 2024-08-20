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

const (
	targetModel = "gemini-1.5-flash"
	prompt      = "Please write a one-page cover letter for the job description and resume."
	resumePath  = "documents/resume.pdf"
)

var logger *slog.Logger

type payload struct {
	JobDescription string `json:"description"`
}

func main() {
	gemApiKey := os.Getenv("GEMINIAPIKEY")
	if gemApiKey == "" {
		log.Fatal("Error. No google gemini API key! Please set env var `export G_API={your key here}`")
	}

	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(gemApiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if false {
		server := http.NewServeMux()
		server.HandleFunc("/gen", genHandler(&ctx, client))
		log.Println("Starting server on 6969")
		log.Fatal(http.ListenAndServe(":6969", server))
	}

	if len(os.Args) <= 1 {
		log.Fatal("Error: please provide a job description.")
	}

	mResp, err := uploaderWrapper(&ctx, client, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	results := toString(mResp)
	fmt.Println(results[0])
}

func genHandler(ctx *context.Context, client *genai.Client) http.HandlerFunc {
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

		mResp, err := uploaderWrapper(ctx, client, msg.JobDescription)
		if err != nil {
			logger.Error("Error", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result := toString(mResp)

		w.Write([]byte(result))
	}
}

func uploaderWrapper(ctx *context.Context, client *genai.Client, desc string) (*genai.GenerateContentResponse, error) {
	resume, err := uploadDocumentFromDisk(ctx, client, resumePath, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.DeleteFile(*ctx, resume.Name)
	logger.Info("uploaded file", "resume", resume)

	combinedPrompt := []genai.Part{
		genai.Text(prompt),
		genai.Text(desc),
		genai.FileData{URI: resume.URI},
	}
	logger.Info("combined prompt", "prompt", combinedPrompt)

	model := client.GenerativeModel(targetModel)
	logger.Info("selected model:", "name", targetModel)

	// Generate content using the prompt.
	mResp, err := model.GenerateContent(*ctx, combinedPrompt...)
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("generated content", "generatedContentResponse", mResp)

	return mResp, nil
}

func toString(mResp *genai.GenerateContentResponse) string {
	var result string
	for _, c := range mResp.Candidates {
		if c.Content != nil {
			result += fmt.Sprint(*c.Content)
		}
	}
	return result
}

// Upload a document useing the file api
func uploadDocumentFromDisk(ctx *context.Context, client *genai.Client, filename string, opts *genai.UploadFileOptions) (*genai.File, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := client.UploadFile(*ctx, "", f, opts)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
