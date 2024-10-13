package command

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/connorkuljis/monorepo/ssg/internal/site"
	"github.com/urfave/cli/v2"
)

var NewPostCommand = cli.Command{
	Name:    "new",
	Aliases: []string{"n"},
	Usage:   "creates a new markdown post",
	Action: func(cCtx *cli.Context) error {
		title := cCtx.Args().First()
		if title == "" {
			return fmt.Errorf("please provide a name")
		}

		page := site.NewBlogPage(title)

		filename := filepath.Join("posts", page.Slug+".md")
		f, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = f.WriteString(page.Matter())
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

		return nil
	},
}
