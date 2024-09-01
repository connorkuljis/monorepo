package templates

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"

	"github.com/labstack/echo/v4"
)

// Implements the `echo.Renderer` interface
type TemplateRegistry struct {
	templates map[string]*template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	_, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("Error, no template found for: %s\n", name)
	}

	log.Println("name:", name)
	log.Println("templates:", t.templates[name])
	return t.templates[name].ExecuteTemplate(w, "base", data)
}

type View struct {
	Name       string
	Base       []string
	Components []string
	View       string
}

func (v View) Parse(fs fs.FS) (*template.Template, error) {
	var ts []string

	ts = append(ts, v.Base...)
	ts = append(ts, v.Components...)
	ts = append(ts, v.View)

	t, err := template.ParseFS(fs, ts...)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewTemplateRegistry(fs fs.FS) (*TemplateRegistry, error) {
	templates := make(map[string]*template.Template)

	base := []string{
		"templates/head.html",
		"templates/base.html",
	}

	views := []View{
		View{
			Name: "index",
			Base: base,
			View: "templates/views/index.html",
		},
		View{
			Name: "upload",
			Base: base,
			View: "templates/views/upload.html",
		},
		View{
			Name: "cover-letter",
			Base: base,
			View: "templates/views/cover-letter.html",
			Components: []string{
				"templates/components/cover-letter.html",
			},
		},
	}

	templates, err := LoadViews(templates, views, fs)
	if err != nil {
		return nil, err
	}

	tr := &TemplateRegistry{
		templates: templates,
	}

	return tr, nil
}

func LoadViews(templates map[string]*template.Template, views []View, fs fs.FS) (map[string]*template.Template, error) {
	for _, v := range views {
		t, err := v.Parse(fs)
		if err != nil {
			return nil, err
		}
		templates[v.Name] = t
	}
	return templates, nil
}
