package site

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"
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

func NewBlogPage(title string) *BlogPage {
	now := time.Now()
	slug := now.Format("2006-01-02") + "-" + util.Slugify(title)
	return &BlogPage{
		Slug:    slug,
		Title:   title,
		Created: now,
	}
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

func (p *BlogPage) Matter() string {
	var builder strings.Builder

	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("title: %s\n", p.Title))
	builder.WriteString(fmt.Sprintf("created: %s\n", p.Created.Format(util.TimeFormat)))
	builder.WriteString("draft: true\n")
	builder.WriteString("---")

	return builder.String()
}
