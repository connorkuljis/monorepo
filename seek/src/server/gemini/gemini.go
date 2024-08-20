package gemini

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	targetModel = "gemini-1.5-flash"
)

type GeminiClient struct {
	Client *genai.Client
	Ctx    *context.Context
	Logger *slog.Logger
	Model  *genai.GenerativeModel
}

func NewGeminiClient(apiKey string, modelName string) (*GeminiClient, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel(modelName)
	// logger.Info("selected model:", "name", targetModel)

	return &GeminiClient{
		Client: client,
		Ctx:    &ctx,
		Logger: logger,
		Model:  model,
	}, nil
}

// f, err := os.Open(filename)
//
//	if err != nil {
//		return nil, err
//	}
//
// defer f.Close()
// defer client.DeleteFile(*ctx, resume.Name)
func (g *GeminiClient) UploadFile(r io.Reader, opts *genai.UploadFileOptions) (*genai.File, error) {
	f, err := g.Client.UploadFile(*g.Ctx, "", r, opts)
	if err != nil {
		return nil, err
	}

	g.Logger.Info("Uploaded file", "name", f.Name, "uri", f.URI)

	return f, nil
}

// Generate content using the prompt.
func (g *GeminiClient) GenContent(prompt []genai.Part) (*genai.GenerateContentResponse, error) {
	resp, err := g.Model.GenerateContent(*g.Ctx, prompt...)
	if err != nil {
		return nil, err
	}

	g.Logger.Info("generated content", "generatedContentResponse", resp)

	return resp, nil
}

func ResumePromptWrapper(jobDescription string, resume *genai.File) []genai.Part {
	defaultPrompt := "Please write a one-page cover letter for the job description and resume."

	return []genai.Part{
		genai.Text(defaultPrompt),
		genai.Text(jobDescription),
		genai.FileData{URI: resume.URI},
	}
}

func ToString(response *genai.GenerateContentResponse) string {
	var result string
	for _, c := range response.Candidates {
		if c.Content != nil {
			result += fmt.Sprint(*c.Content)
		}
	}
	return result
}
