package main

import (
	"fmt"
	"log"
	"os"

	"github.com/connorkuljis/seek-js/gemini"
)

func main() {
	gemApiKey := os.Getenv("GEMINIAPIKEY")
	if gemApiKey == "" {
		log.Fatal("Error. No google gemini API key! Please set env var `export G_API={your key here}`")
	}

	if len(os.Args) <= 1 {
		log.Fatal("Error not enough arguments")
	}

	g, err := gemini.NewGeminiClient(gemApiKey, "gemini-1.5-flash")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("static/Connor-Kuljis_Resume_2024-07.pdf")
	if err != nil {
		log.Fatal(err)
	}

	gf, err := g.UploadFile(f, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Client.DeleteFile(*g.Ctx, gf.Name)

	p := gemini.ResumePromptWrapper(os.Args[1], gf)

	resp, err := g.GenerateContent(p)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(gemini.ToString(resp))
}
