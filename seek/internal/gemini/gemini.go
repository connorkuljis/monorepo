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

func NewClient(ctx *context.Context, apiKey string, logger *slog.Logger) (*GeminiClient, error) {
	client, err := genai.NewClient(*ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	logger.Info("Created Gemini Client")

	return &GeminiClient{
		Client: client,
		Ctx:    ctx,
		Logger: logger,
	}, nil
}

func (gc *GeminiClient) UploadFile(r io.Reader, filename string, opts *genai.UploadFileOptions) (*genai.File, error) {
	f, err := gc.Client.UploadFile(*gc.Ctx, filename, r, opts)
	if err != nil {
		return nil, err
	}

	gc.Logger.Info("Uploaded file", "name", f.Name, "uri", f.URI)

	return f, nil
}

// Generate content using the prompt.
func (g *GeminiClient) GenerateContent(prompt []genai.Part, model string) (*genai.GenerateContentResponse, error) {
	m := g.Client.GenerativeModel(model)

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
