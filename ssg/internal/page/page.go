package page

import (
	"errors"
	"html/template"
	"io"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/russross/blackfriday/v2"
)

const (
	root            = "templates/root.html"
	layout          = "templates/layout.html"
	head            = "templates/head.html"
	componentHeader = "templates/components/header.html"
	viewPost        = "templates/views/post.html"
	viewIndex       = "templates/views/index.html"
)

type HomePage struct {
	Filename      string
	FeaturedPosts []*BlogPage
}

type BlogPage struct {
	Length    int
	Preview   string
	Filename  string
	Content   template.HTML
	Templates []string

	Title   string    `yaml:"title"`
	Created time.Time `yaml:"created"`
	Tags    []string  `yaml:"tags"`
	Draft   bool      `yaml:"draft"`
}

func ParsePage(r io.Reader) (*BlogPage, error) {
	var page BlogPage
	content, err := frontmatter.Parse(r, &page)
	if err != nil {
		return &BlogPage{}, err
	}

	err = page.Validate()
	if err != nil {
		return &BlogPage{}, err
	}

	page.Length = len(content)
	page.Preview = string(content[:100])
	page.Filename = page.Title + ".html"
	page.Content = template.HTML(blackfriday.Run(content))

	return &page, nil
}

func (p *BlogPage) Generate(w io.Writer) error {
	t, err := template.New(p.Filename).ParseFiles(root, layout, head, componentHeader, viewPost)
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(w, "root", map[string]any{"Post": p})
	if err != nil {
		return err
	}

	return nil
}

func (p *HomePage) Generate(w io.Writer) error {
	t, err := template.New("index.html").ParseFiles(root, layout, head, componentHeader, viewIndex)
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(w, "root", map[string]any{"Posts": p.FeaturedPosts})
	if err != nil {
		return err
	}

	return nil
}

func (p *BlogPage) Validate() error {
	if p.Title == "" {
		return errors.New("Invalid post matter. Title is empty")
	}

	if p.Created.IsZero() {
		return errors.New("Date is zero.")
	}

	return nil
}
