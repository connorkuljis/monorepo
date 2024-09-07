package gemini

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	Client *genai.Client
	Ctx    *context.Context
	Logger *slog.Logger
}

type Model int

const (
	Flash Model = iota
	Pro
)

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

func (g *GeminiClient) UploadFile(r io.Reader, filename string, opts *genai.UploadFileOptions) (*genai.File, error) {
	f, err := g.Client.UploadFile(*g.Ctx, filename, r, opts)
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

	m.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := m.GenerateContent(*g.Ctx, prompt...)
	if err != nil {
		return nil, err
	}

	g.Logger.Info("generated content", "total_token_count", resp.UsageMetadata.TotalTokenCount)

	return resp, nil
}

func ToString(resp *genai.GenerateContentResponse) string {
	var result string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			result += fmt.Sprint(p)
		}
	}

	return result
}
