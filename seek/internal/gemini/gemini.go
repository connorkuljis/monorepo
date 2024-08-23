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

type GeminiClient struct {
	Client *genai.Client
	Ctx    *context.Context
	Logger *slog.Logger
	Model  *genai.GenerativeModel
}

func NewGeminiClient(apiKey string, targetModel string) (*GeminiClient, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel(targetModel)

	return &GeminiClient{
		Client: client,
		Ctx:    &ctx,
		Logger: logger,
		Model:  model,
	}, nil
}

func (g *GeminiClient) UploadFile(r io.Reader, opts *genai.UploadFileOptions) (*genai.File, error) {
	f, err := g.Client.UploadFile(*g.Ctx, "", r, opts)
	if err != nil {
		return nil, err
	}

	g.Logger.Info("Uploaded file", "name", f.Name, "uri", f.URI)

	return f, nil
}

// Generate content using the prompt.
func (g *GeminiClient) GenerateContent(prompt []genai.Part) (*genai.GenerateContentResponse, error) {
	resp, err := g.Model.GenerateContent(*g.Ctx, prompt...)
	if err != nil {
		return nil, err
	}

	g.Logger.Info("generated content", "generatedContentResponse", resp)

	return resp, nil
}

func ResumePromptWrapper(jobDescription string, uri string) []genai.Part {
	defaultPrompt := "Please write a one-page cover letter for the job description and resume."

	return []genai.Part{
		genai.Text(defaultPrompt),
		genai.Text(jobDescription),
		genai.FileData{URI: uri},
	}
}

func ToString(response *genai.GenerateContentResponse) string {
	var result string
	for _, c := range response.Candidates {
		for _, p := range c.Content.Parts {
			result += fmt.Sprint(p)
		}
		// if c.Content != nil {
		// 	result += fmt.Sprint(*c.Content)
		// }
	}
	return result
}

// TODO:
// - pass logger as dependency to NewGeminiClient OR functional options builder pattern `withLogger(logger logger)`
// - implement delete file `defer client.DeleteFile(*ctx, resume.Name)``