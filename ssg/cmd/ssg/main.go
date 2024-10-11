package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/connorkuljis/monorepo/ssg/internal/site"
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
			// {
			// 	Name:    "new",
			// 	Aliases: []string{"n"},
			// 	Usage:   "creates a new markdown post",
			// 	Action: func(cCtx *cli.Context) error {
			// 		NewPostCommand()
			// 		return nil
			// 	},
			// },
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func BuildBlogCommand(includeDrafts bool) error {
	fmt.Println("Initialising site...")
	site, err := site.NewSite(includeDrafts)
	if err != nil {
		return err
	}

	err = site.CreateNewPublicDir()
	if err != nil {
		return err
	}

	err = site.BundleStaticContentToPublicDir()
	if err != nil {
		return err
	}

	fmt.Println("Parsing posts...")
	err = site.ParseMarkdownPosts()
	if err != nil {
		return err
	}
	fmt.Println("Built", len(site.BlogPages), "posts")

	fmt.Println("Building home page...")
	err = site.BuildHomePage()
	if err != nil {
		return err
	}

	fmt.Println("Saving blog...")
	n, err := site.Generate()
	if err != nil {
		return err
	}
	fmt.Println("Published", n, "/", len(site.BlogPages), "posts")

	fmt.Println("Done!")
	return nil
}

// func NewPostCommand() {
// 	if len(os.Args) <= 2 {
// 		log.Fatal(errors.New("missing argument, please provide a title."))
// 	}

// 	name := os.Args[2]
// 	name = util.Slugify(name)
// 	filename := filepath.Join("posts", name+".md")
// 	f, err := os.Create(filename)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer f.Close()

// 	matter := matter.Matter{
// 		Title:   name,
// 		Created: time.Now(),
// 	}

// 	header := `
// ---
// name: %s
// date: %s
// draft: %s
// ---
// `
// 	s := fmt.Sprintf(
// 		header,
// 		matter.Title,
// 		matter.Created.Format(TimeFormat),
// 		"true",
// 	)

// 	_, err = f.WriteString(s)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// editor env var
// 	editor := os.Getenv("EDITOR")
// 	if editor == "" {
// 		log.Fatal("Error: $EDITOR not set!")
// 	}

// 	// open file with editor
// 	cmd := exec.Command(editor, filename)

// 	// echo exec command
// 	fmt.Println("exec:", cmd.String())

// 	// set the process output to the os output
// 	cmd.Stdout = os.Stdout

// 	// run the command
// 	if err = cmd.Run(); err != nil {
// 		log.Println(err)
// 	}
// }

func ServeCommand() {
	http.Handle("/", http.FileServer(http.Dir("public")))
	fmt.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
