package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/labstack/echo/v4"
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
	JobTitle          string `json:"job_title"`
	Introduction      string `json:"introduction"`
	Body              struct {
		Qualification1 string   `json:"qualification_1"`
		Qualification2 string   `json:"qualification_2"`
		Qualification3 string   `json:"qualification_3"`
		Experience1    string   `json:"experience_1"`
		Experience2    string   `json:"experience_2"`
		Experience3    string   `json:"experience_3"`
		Skills         []string `json:"skills"`
	} `json:"body"`
	Closing string `json:"closing"`

	Phone              string
	Email              string
	Filename           string
	Date               time.Time
	ListingDescription string
	ListingOrigin      string
	Prompt             string
	ResumeURI          string
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

func NewCoverLetter(resumeURI, filename, email, phone, listingDescription, listingOrigin string, date time.Time) *CoverLetter {
	prompt := `Please write a cover letter using this JSON schema:
{
  "type": "object",
  "properties": {
    "applicant_full_name": { "type": "string" },
    "company_name": { "type": "string" },
    "job_title": { "type": "string" },
    "introduction": { "type": "string" },
    "body": {
      "type": "object",
      "properties": {
        "qualification_1": { "type": "string" },
        "qualification_2": { "type": "string" },
        "qualification_3": { "type": "string" },
        "experience_1": { "type": "string" },
        "experience_2": { "type": "string" },
        "experience_3": { "type": "string" },
        "skills": { "type": "array", "items": { "type": "string" } }
      }
    },
    "closing": { "type": "string" }
  }
}

Your responses should always be complete and only include regular characters.

Under no circumstance should your responses include any placeholders. It is getting send directly to the recruiter. 
`

	return &CoverLetter{
		Phone:              phone,
		Email:              email,
		Filename:           filename,
		Date:               date,
		ListingDescription: listingDescription,
		ListingOrigin:      listingOrigin,
		Prompt:             prompt,
		ResumeURI:          resumeURI,
	}
}

// Generate content using the prompt.
func (cv *CoverLetter) Generate(gemini *GeminiClient, modelName string) error {
	parts := []genai.Part{
		genai.Text(cv.Prompt),
		genai.Text("Place where I found the job description: " + cv.ListingOrigin),
		genai.Text(cv.ListingDescription),
		// genai.FileData{URI: uri},
	}

	genaiModel := gemini.Client.GenerativeModel(modelName)
	genaiModel.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	resp, err := genaiModel.GenerateContent(*gemini.Ctx, parts...)
	if err != nil {
		return err
	}

	var result string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			result += fmt.Sprint(p)
		}
	}

	err = json.Unmarshal([]byte(result), cv)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return err
	}

	return nil
}

func (cv *CoverLetter) Render(c echo.Context) ([]byte, error) {
	data := map[string]any{
		"CoverLetter": cv,
	}
	var buf bytes.Buffer
	renderer := c.Echo().Renderer
	err := renderer.Render(&buf, "cover-letter-print.html", data, c)
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

// func (cv *CoverLetter) Save() error {
// 	path := filepath.Join("out", cv.Filename)
// 	err := os.MkdirAll(path, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}

// 	filename := filepath.Join(path, "index.html")
// 	f, err := os.Create(filename)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = f.Write(html)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
