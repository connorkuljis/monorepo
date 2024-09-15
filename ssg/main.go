package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/russross/blackfriday/v2"
)

type Matter struct {
	Name string    `yaml:"name"`
	Date time.Time `yaml:"date"`
	Tags []string  `yaml:"tags"`
}

type Post struct {
	Template *template.Template
	Matter   Matter
	Body     template.HTML
	Filename string
}

func (m *Matter) Validate() error {
	if m.Name == "" {
		return errors.New("Mising matter name.")
	}

	if m.Date.IsZero() {
		return errors.New("Date is zero.")
	}

	return nil
}

func main() {
	log.Println(time.Now().Format(time.RFC3339))
	serve := flag.Bool("serve", false, "serve the public files, if not passed by default build is run")
	flag.Parse()

	if *serve {
		Serve()
		return
	}

	build()
}

func build() {
	posts, err := ParsePosts()
	if err != nil {
		log.Fatal(err)
	}
	MakePublic(posts)
}

func ParsePosts() ([]Post, error) {
	var posts []Post

	// we need to read all the fds from the posts directory
	files, err := os.ReadDir("posts")
	if err != nil {
		return posts, err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".md" {

			target := filepath.Join("posts", file.Name())
			b, err := os.ReadFile(target)
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}

			// Matter stuff
			var matter Matter
			body, err := frontmatter.Parse(strings.NewReader(string(b)), &matter)
			if err != nil {
				fmt.Println("Error parsing matter:", err.Error())
				continue
			}

			err = matter.Validate()
			if err != nil {
				fmt.Println(err)
				continue
			}

			p := Post{
				Matter:   matter,
				Body:     template.HTML(blackfriday.Run(body)),
				Filename: matter.Name + ".html",
			}

			p.Template, err = template.New(p.Filename).ParseFiles("templates/base.html", "templates/post.html")
			if err != nil {
				fmt.Println("Error parsing template:", err)
				continue
			}

			posts = append(posts, p)
		}
	}

	return posts, nil
}

func MakePublic(posts []Post) error {
	// create base directory
	base := "public"
	err := os.Mkdir(base, os.ModePerm)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			fmt.Println(err.Error())
		default:
			log.Fatal("error creating path")
		}
	}

	t, err := template.New("index.html").ParseFiles("templates/base.html", "templates/index.html")
	if err != nil {
		log.Fatal(err)

	}
	f, err := os.Create(filepath.Join(base, t.Name()))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = t.ExecuteTemplate(f, "base", map[string]any{"Posts": posts})
	if err != nil {
		log.Fatal(err)
	}

	postsD := filepath.Join(base, "posts")
	err = os.Mkdir(postsD, os.ModePerm)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			fmt.Println(err.Error())
		default:
			log.Fatal("error creating path")
		}
	}

	// for each template, create a file from the template name and
	for _, post := range posts {
		target := filepath.Join(postsD, post.Filename)
		f, err := os.Create(target)
		if err != nil {
			fmt.Println("Error creating file", f.Name())
			continue
		}
		defer f.Close()
		post.Template.ExecuteTemplate(f, "base", post)
	}

	return nil
}

func Serve() {
	http.Handle("/", http.FileServer(http.Dir("public")))
	fmt.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
