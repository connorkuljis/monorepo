package main

import (
	"log"
	"os"

	"github.com/connorkuljis/seek-js/gemini"
)

func main() {
	gemApiKey := os.Getenv("GEMINIAPIKEY")
	if gemApiKey == "" {
		log.Fatal("Error. No google gemini API key! Please set env var `export G_API={your key here}`")
	}

	_, err := gemini.NewGeminiClient(gemApiKey, "gemini-1.5-flash")
	if err != nil {
		log.Fatal(err)
	}
}
