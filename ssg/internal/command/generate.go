package command

import (
	"fmt"

	"github.com/connorkuljis/monorepo/ssg/internal/site"
	"github.com/urfave/cli/v2"
)

var GenerateCommand = cli.Command{
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

		fmt.Println("Initialising site...")
		site := site.NewSite(includeDrafts)

		// create the skeleton
		err := site.CreateNewPublicDir()
		if err != nil {
			return err
		}

		err = site.BundleStaticContentToPublicDir()
		if err != nil {
			return err
		}

		fmt.Println("Loading blog pages...")
		err = site.LoadAllBlogPages()
		if err != nil {
			return err
		}

		fmt.Println("Building home page...")
		err = site.BuildHomePage()
		if err != nil {
			return err
		}

		fmt.Println("Saving blog...")
		err = site.Generate()
		if err != nil {
			return err
		}

		fmt.Println("Done!")

		return nil
	},
}
