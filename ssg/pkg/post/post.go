package post

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/connorkuljis/monorepo/ssg/pkg/matter"
	"github.com/russross/blackfriday/v2"
)

const base = "base"

type Post struct {
	Length  int
	Preview string
	Dist    string

	HTML   template.HTML
	Matter matter.Matter
}

func NewPost(content []byte) (Post, error) {
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
		Length:  len(rest),
		Preview: string(rest[:100]),
		Dist:    matter.Name + ".html",
		HTML:    template.HTML(blackfriday.Run(rest)),
		Matter:  matter,
	}

	return p, nil
}

func (p *Post) Render(path string) error {
	f, err := os.Create(filepath.Join(path, p.Dist))
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := template.New(p.Dist).ParseFiles("templates/base.html", "templates/post.html")
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(f, base, p)
	if err != nil {
		return err
	}

	return nil
}

func (p *Post) PPrint() {
	fmt.Println(p.Dist)
	fmt.Println("\tName:", p.Matter.Name)
	fmt.Println("\tDate:", p.Matter.Date)
	fmt.Println("\tDraft:", p.Matter.Draft)
	fmt.Println("\tLength:", p.Length)
}
