package cv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
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

Under no circumstance should your responses include any placeholders, for example: 
	- "[Platform where you found the job]" or 
	- "[Your Email Address]"

If you do not know, skip it.
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

func (cv *CoverLetter) SaveAsHTML() error {
	t, err := template.New("").ParseFiles("templates/partial-cover-letter.html", "templates/head.html", "templates/component-cover-letter.html")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "partial-cover-letter", cv)
	if err != nil {
		return err
	}

	path := filepath.Join("out", cv.Filename)
	err = os.MkdirAll(path, os.ModePerm)

	filename := filepath.Join(path, "index.html")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
