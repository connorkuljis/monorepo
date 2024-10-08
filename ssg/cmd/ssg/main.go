package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/connorkuljis/monorepo/ssg/internal/blog"
	"github.com/connorkuljis/monorepo/ssg/internal/matter"
	"github.com/urfave/cli/v2"
)

const (
	TimeFormat = time.RFC3339
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "generate",
				Aliases: []string{"gen", "g"},
				Usage:   "generates site html and css into `/public`",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "include-drafts",
						Usage:    "include drafts",
						Required: false,
						Aliases:  []string{"d"},
					},
				},
				Action: func(cCtx *cli.Context) error {
					includeDrafts := cCtx.Bool("include-drafts")
					err := BuildBlogCommand(includeDrafts)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"server", "s"},
				Usage:   "serves the static content in `/public`",
				Action: func(cCtx *cli.Context) error {
					ServeCommand()
					return nil
				},
			},
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "creates a new markdown post",
				Action: func(cCtx *cli.Context) error {
					NewPostCommand()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func BuildBlogCommand(includeDrafts bool) error {
	blog, err := blog.NewBlog(includeDrafts)
	if err != nil {
		return err
	}

	fmt.Println("Initialising blog...")
	err = blog.Init()
	if err != nil {
		return err
	}

	fmt.Println("Building posts...")
	err = blog.BuildPosts()
	if err != nil {
		return err
	}
	fmt.Println("Built", len(blog.Posts), "posts")

	fmt.Println("Building home page...")
	err = blog.BuildHomePage()
	if err != nil {
		return err
	}

	fmt.Println("Saving blog...")
	n, err := blog.Save()
	if err != nil {
		return err
	}
	fmt.Println("Published", n, "/", len(blog.Posts), "posts")

	fmt.Println("Done!")
	return nil
}

func NewPostCommand() {
	if len(os.Args) <= 2 {
		log.Fatal(errors.New("missing argument, please provide a title."))
	}

	name := os.Args[2]
	filename := filepath.Join("posts", name+".md")
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	matter := matter.Matter{
		Name: name,
		Date: time.Now(),
	}

	header := `
---
name: %s
date: %s
draft: %s
---
`
	s := fmt.Sprintf(
		header,
		matter.Name,
		matter.Date.Format(TimeFormat),
		"true",
	)

	_, err = f.WriteString(s)
	if err != nil {
		log.Fatal(err)
	}

	// editor env var
	editor := os.Getenv("EDITOR")
	if editor == "" {
		log.Fatal("Error: $EDITOR not set!")
	}

	// open file with editor
	cmd := exec.Command(editor, filename)

	// echo exec command
	fmt.Println("exec:", cmd.String())

	// set the process output to the os output
	cmd.Stdout = os.Stdout

	// run the command
	if err = cmd.Run(); err != nil {
		log.Println(err)
	}
}

func ServeCommand() {
	http.Handle("/", http.FileServer(http.Dir("public")))
	fmt.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
