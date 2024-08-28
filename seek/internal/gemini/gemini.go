package gemini

import (
	"context"
	"encoding/json"
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

type CoverLetter struct {
	ApplicantFullName string `json:"applicant_full_name"`
	CompanyName       string `json:"company_name"`
	Introduction      string `json:"introduction"`
	Body              string `json:"body"`
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

func ResumePromptWrapper(jobDescription string, uri string) []genai.Part {
	prompt := `Please write a cover letter using this JSON schema:
        { "type": "object",
          "properties": {
              "applicant_full_name": { "type": "string" },
              "company_name": { "type": "string" },
              "introduction": { "type": "string" },
              "body": { "type": "string" },
         }
        }`

	parts := []genai.Part{
		genai.Text(prompt),
		genai.Text(jobDescription),
		genai.FileData{URI: uri},
	}

	return parts
}

func ToString(resp *genai.GenerateContentResponse) string {
	var result string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			result += fmt.Sprint(p)
		}
	}

	// for _, c := range resp.Candidates {
	// 	if c.Content != nil {
	// 		// fmt.Println(*c.Content)
	// 		result += fmt.Sprint(*c.Content)
	// 	}
	// }

	return result
}

func NewCoverLetterFromJSON(data string) (CoverLetter, error) {
	var cv CoverLetter
	err := json.Unmarshal([]byte(data), &cv)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return cv, err
	}

	// Access and print the parsed values
	fmt.Println("Applicant Full Name:", cv.ApplicantFullName)
	fmt.Println("Company Name:", cv.CompanyName)
	fmt.Println("Introduction:", cv.Introduction)
	fmt.Println("Body:", cv.Body)

	return cv, nil

}

// TODO:
// - pass logger as dependency to NewGeminiClient OR functional options builder pattern `withLogger(logger logger)`
// - implement delete file `defer client.DeleteFile(*ctx, resume.Name)``
