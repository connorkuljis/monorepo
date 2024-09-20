package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/connorkuljis/monorepo/ssg/pkg/blog"
	"github.com/connorkuljis/monorepo/ssg/pkg/matter"
)

const (
	TimeFormat = time.RFC3339
	PublicDir  = "public"
	PostsDir   = "posts"
	SourceDir  = "posts"
)

func main() {
	serve := flag.Bool("serve", false, "serve the public files, if not passed by default build is run")
	new := flag.Bool("new", false, "make a new post")
	flag.Parse()

	if *serve {
		ServeCommand()
		return
	} else if *new {
		NewPostCommand()
		return
	} else {
		BuildBlogCommand()
		return
	}
}

func BuildBlogCommand() error {
	blog, err := blog.NewBlog()
	if err != nil {
		return err
	}

	err = blog.Init()
	if err != nil {
		log.Fatal("error:", err.Error())
	}

	err = blog.BuildPosts()
	if err != nil {
		log.Fatal("error:", err.Error())
	}

	err = blog.BuildHomePage()
	if err != nil {
		log.Fatal("error:", err.Error())
	}

	err = blog.Save()
	if err != nil {
		log.Fatal("error:", err.Error())
	}

	return nil
}

func NewPostCommand() {
	if len(os.Args) <= 2 {
		log.Fatal(errors.New("missing name"))
	}

	name := os.Args[2]
	f, err := os.Create(filepath.Join("posts", name+".md"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	matter := matter.Matter{
		Name: name,
		Date: time.Now(),
	}

	s := fmt.Sprintf(`
---
name: %s
date: %s
draft: %s
---
`,
		matter.Name,
		matter.Date.Format(TimeFormat),
		"true",
	)

	_, err = f.WriteString(s)
	if err != nil {
		log.Fatal(err)
	}
}

func ServeCommand() {
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
