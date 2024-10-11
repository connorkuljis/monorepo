package site

import (
	"html/template"
	"io"

	"github.com/connorkuljis/monorepo/ssg/internal/util"
)

type HomePage struct {
	Filename      string
	FeaturedPosts []*BlogPage
}

func (p *HomePage) Generate(w io.Writer) error {
	t, err := template.New(p.Filename).ParseFiles(util.HomePageTemplates()...)
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(w, "root", map[string]any{"Posts": p.FeaturedPosts})
	if err != nil {
		return err
	}

	return nil
}
