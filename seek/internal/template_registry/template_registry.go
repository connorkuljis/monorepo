package template_registry

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

const PartialIdentifier = "partial-"

// Implements the `echo.Renderer` interface
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// this function provides a mechanism for rendering both full templates and partial templates based on their names.
// The PartialIdentifier constant allows for flexible identification of partial templates, while the apply variable ensures that the correct target template is used for rendering.
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	apply := "base"

	if strings.HasPrefix(name, PartialIdentifier) {
		apply = strings.TrimPrefix(name, PartialIdentifier)
	}

	_, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("Error, no template found for: %s\n", name)
	}

	return t.templates[name].ExecuteTemplate(w, apply, data)
}

func NewTemplateRegistry(fs fs.FS, templateDir string) (*TemplateRegistry, error) {
	templates := make(map[string]*template.Template)

	base := []string{
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, "head.html"),
		filepath.Join(templateDir, "layout.html"),
	}

	views := []View{
		View{
			Name: "index",
			Base: base,
			View: filepath.Join(templateDir, "views/index.html"),
		},
		View{
			Name: "upload",
			Base: base,
			View: filepath.Join(templateDir, "views/upload.html"),
		},
		View{
			Name: "cover-letter",
			Base: base,
			View: filepath.Join(templateDir, "views/cover-letter.html"),
			Components: []string{
				filepath.Join(templateDir, "components/cover-letter.html"),
			},
		},
		View{
			Name: "cover-letter-print",
			Base: []string{
				filepath.Join(templateDir, "base.html"),
				filepath.Join(templateDir, "head.html"),
				filepath.Join(templateDir, "layout-print.html"),
			},
			View: filepath.Join(templateDir, "views/cover-letter-print.html"),
			Components: []string{
				filepath.Join(templateDir, "components/cover-letter.html"),
			},
		},
	}

	partials := []Partial{
		Partial{
			Name: "partial-foo",
			Components: []string{
				filepath.Join(templateDir, "components/foo.html"),
				filepath.Join(templateDir, "components/bar.html"),
			},
		},
		Partial{
			Name: "partial-bar",
			Components: []string{
				filepath.Join(templateDir, "components/bar.html"),
			},
		},
	}

	templates, err := LoadViews(templates, views, fs)
	if err != nil {
		return nil, err
	}

	templates, err = LoadPartials(templates, partials, fs)
	if err != nil {
		return nil, err
	}

	tr := &TemplateRegistry{
		templates: templates,
	}

	return tr, nil
}
