package post

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/connorkuljis/monorepo/ssg/internal/matter"
	"github.com/russross/blackfriday/v2"
)

const root = "root"

type Post struct {
	Length   int
	Preview  string
	Filename string
	HTML     template.HTML
	Matter   matter.Matter
}

func ParsePost(content []byte) (Post, error) {
	var matter matter.Matter
	rest, err := frontmatter.Parse(strings.NewReader(string(content)), &matter)
	if err != nil {
		return Post{}, err
	}

	err = matter.Validate()
	if err != nil {
		return Post{}, err
	}

	p := Post{
		Length:   len(rest),
		Preview:  string(rest[:100]),
		Filename: matter.Name + ".html",
		HTML:     template.HTML(blackfriday.Run(rest)),
		Matter:   matter,
	}

	return p, nil
}

func (p *Post) Render(path string) error {
	f, err := os.Create(filepath.Join(path, p.Filename))
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := template.New(p.Filename).ParseFiles(
		"templates/layout.html",
		"templates/head.html",
		"templates/view/post.html",
	)
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(f, root, p)
	if err != nil {
		return err
	}

	return nil
}

func (p *Post) PPrint() {
	fmt.Println(p.Filename)
	fmt.Println("\tName:", p.Matter.Name)
	fmt.Println("\tDate:", p.Matter.Date)
	fmt.Println("\tDraft:", p.Matter.Draft)
	fmt.Println("\tLength:", p.Length)
}
