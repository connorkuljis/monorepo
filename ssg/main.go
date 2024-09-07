package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/russross/blackfriday/v2"
)

type matter struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
}

func main() {
	t, err := ConvertPostsToHTML()
	if err != nil {
		log.Fatal(err)
	}
	Save(t)
	// Serve()
}

func ConvertPostsToHTML() ([]*template.Template, error) {
	files, err := os.ReadDir("posts")
	if err != nil {
		return nil, err
	}

	var tpls []*template.Template
	base := `{{ define "base" }}<html><body>{{ template "post" . }}</body></html>{{ end }}`
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {
			markdownBytes, err := os.ReadFile(filepath.Join("posts", file.Name()))
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}

			var matter matter
			rest, err := frontmatter.Parse(strings.NewReader(string(markdownBytes)), &matter)
			if err != nil {
				fmt.Println("Error parsing matter:", err)
				continue
			}
			log.Println(matter)

			name := matter.Name
			if name == "" {
				name = file.Name()
			}

			htmlBytes := fmt.Sprintf(base+`{{ define "post" }}%s{{ end }}`, blackfriday.Run(rest))
			t, err := template.New(name).Parse(string(htmlBytes))
			if err != nil {
				fmt.Println("Error parsing template:", err)
				continue
			}
			w := os.Stdout
			t.ExecuteTemplate(w, "base", nil)

			tpls = append(tpls, t)
		}
	}

	return tpls, nil
}

func Save(tpls []*template.Template) error {
	err := os.Mkdir("out", os.ModePerm)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			fmt.Println(err.Error())
		default:
			log.Fatal("error creating path")
		}
	}
	for _, t := range tpls {
		f, err := os.Create(filepath.Join("out", t.Name()+".html"))
		if err != nil {
			fmt.Println("Error creating file", f.Name())
			continue
		}
		defer f.Close()

		t.ExecuteTemplate(f, "base", nil)
		fmt.Println("Created file:", f.Name())
	}
	return nil
}

func Serve() {
	http.Handle("/", http.FileServer(http.Dir("out")))
	fmt.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
