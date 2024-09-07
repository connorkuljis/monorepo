package cv

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

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

	Phone    string
	Email    string
	Filename string
	Date     time.Time
}

const (
	Prompt = `Please write a cover letter using this JSON schema:
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
)

func NewCoverLetterFromJSON(filename, jsonString string) (CoverLetter, error) {
	var cv CoverLetter
	err := json.Unmarshal([]byte(jsonString), &cv)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return cv, err
	}

	cv.Filename = filename
	cv.Date = time.Now()

	return cv, nil
}

func (cv *CoverLetter) SaveAsHTML(html []byte) error {
	path := filepath.Join("out", cv.Filename)

	err := os.MkdirAll(path, os.ModePerm)

	filename := filepath.Join(path, "index.html")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = f.Write(html)
	if err != nil {
		return err
	}

	return nil
}
