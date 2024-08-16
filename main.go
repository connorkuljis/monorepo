package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	targetModel = "gemini-1.5-flash"
	// targetModel                 = "gemini-1.5-pro"
	resumePath                  = "documents/resume.pdf"
	promptToGenerateCoverLetter = "Please write a one-page cover letter for the job description and resume."
)

func main() {
	apiKey := os.Getenv("G_API")
	if apiKey == "" {
		log.Fatal("Error. No google API key!")
	}

	if len(os.Args) <= 1 {
		log.Fatal(errors.New("No input provided"))
	}
	desc := os.Args[1]

	// Initialise the model
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel(targetModel)

	doc, err := uploadDocument(ctx, client, resumePath, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Uploaded file %s as: %q\n", doc.DisplayName, doc.URI)

	// Create a prompt using text and the URI reference for the uploaded file.
	prompt := []genai.Part{
		genai.Text(promptToGenerateCoverLetter),
		genai.Text(desc),
		genai.FileData{URI: doc.URI},
	}

	// Generate content using the prompt.
	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		log.Fatal(err)
	}

	// Handle the response of generated text
	for _, c := range resp.Candidates {
		if c.Content != nil {
			fmt.Println(*c.Content)
		}
	}
}

// Upload a document useing the file api
func uploadDocument(ctx context.Context, client *genai.Client, filename string, opts *genai.UploadFileOptions) (*genai.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc, err := client.UploadFile(ctx, "", file, opts)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
