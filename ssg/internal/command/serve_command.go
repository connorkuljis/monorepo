package command

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

var ServeCommand = cli.Command{
	Name:    "serve",
	Aliases: []string{"server", "s"},
	Usage:   "serves the static content in `/public`",
	Action: func(cCtx *cli.Context) error {
		http.Handle("/", http.FileServer(http.Dir("public")))
		fmt.Println("Server listening on port 8080")
		return http.ListenAndServe(":8080", nil)
	},
}
