package site

import (
	"errors"
	"html/template"
	"io"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/connorkuljis/monorepo/ssg/internal/util"
	"github.com/russross/blackfriday/v2"
)

type BlogPage struct {
	Slug    string
	Content template.HTML

	Title   string    `yaml:"title"`
	Created time.Time `yaml:"created"`
	Tags    []string  `yaml:"tags"`
	Draft   bool      `yaml:"draft"`
}

func ParsePage(r io.Reader) (*BlogPage, error) {
	// marshall the front matter
	var page BlogPage
	contentBytes, err := frontmatter.Parse(r, &page)
	if err != nil {
		return &BlogPage{}, err
	}

	err = page.Validate()
	if err != nil {
		return &BlogPage{}, err
	}

	// create the slug
	timestamp := page.Created.Format("2006-01-02")
	slug := util.Slugify(timestamp+"-"+page.Title) + ".html"

	// parse raw markdown content to html
	contentHTML := template.HTML(blackfriday.Run(contentBytes))

	page.Slug = slug
	page.Content = contentHTML

	return &page, nil
}

func (p *BlogPage) Generate(w io.Writer) error {
	t, err := template.New(p.Slug).ParseFiles(util.BlogPageTemplates()...)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, "root", map[string]any{"Post": p})
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
