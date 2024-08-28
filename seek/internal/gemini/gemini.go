package gemini

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Model int

const (
	Flash Model = iota
	Pro
)

type GeminiClient struct {
	Client *genai.Client
	Ctx    *context.Context
	Logger *slog.Logger
}

func NewGeminiClient(apiKey string, logger *slog.Logger) (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	g := &GeminiClient{
		Client: client,
		Ctx:    &ctx,
		Logger: logger,
	}

	return g, nil
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
func (g *GeminiClient) GenerateContent(prompt []genai.Part, model Model) (*genai.GenerateContentResponse, error) {
	var name string
	switch model {
	case Flash:
		name = "gemini-1.5-flash"
	case Pro:
		name = "gemini-1.5-pro"
	}

	m := g.Client.GenerativeModel(name)

	resp, err := m.GenerateContent(*g.Ctx, prompt...)
	if err != nil {
		return nil, err
	}

	g.Logger.Info("generated content", "total_token_count", resp.UsageMetadata.TotalTokenCount)

	return resp, nil
}

func ResumePromptWrapper(jobDescription string, uri string) []genai.Part {
	defaultPrompt := "Please write a one-page cover letter for the job description and resume."

	parts := []genai.Part{
		genai.Text(defaultPrompt),
		genai.Text(jobDescription),
		genai.FileData{URI: uri},
	}

	return parts
}

func ToString(response *genai.GenerateContentResponse) string {
	var result string
	for _, c := range response.Candidates {
		for _, p := range c.Content.Parts {
			result += fmt.Sprint(p)
		}
	}
	return result
}

// TODO:
// - pass logger as dependency to NewGeminiClient OR functional options builder pattern `withLogger(logger logger)`
// - implement delete file `defer client.DeleteFile(*ctx, resume.Name)``
