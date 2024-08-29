package cv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type CoverLetter struct {
	ApplicantFullName string `json:"applicant_full_name"`
	CompanyName       string `json:"company_name"`
	Introduction      string `json:"introduction"`
	Body              string `json:"body"`

	Filename string
}

const (
	Prompt = `Please write a cover letter using this JSON schema:
        { "type": "object",
          "properties": {
              "applicant_full_name": { "type": "string" },
              "company_name": { "type": "string" },
              "introduction": { "type": "string" },
              "body": { "type": "string" },
         }
        }`
)

func NewCoverLetterFromJSON(filename, jsonString string) (CoverLetter, error) {
	var cv CoverLetter
	err := json.Unmarshal([]byte(jsonString), &cv)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return cv, err
	}
	cv.Filename = filename
	return cv, nil
}

func (cv *CoverLetter) SaveAsHTML() error {
	t, err := template.New("").ParseFiles("templates/cover-letter.html", "templates/head.html")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, "cover-letter", cv)
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
