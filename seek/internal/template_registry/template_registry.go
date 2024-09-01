package template_registry

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"

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

	return t.templates[name].ExecuteTemplate(w, "base", data)
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

	templates, err := LoadViews(templates, views, fs)
	if err != nil {
		return nil, err
	}

	tr := &TemplateRegistry{
		templates: templates,
	}

	return tr, nil
}
