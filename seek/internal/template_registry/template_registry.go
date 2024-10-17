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

const (
	DefaultRenderTarget = "base"
	PartialIdentifier   = "partial-"
)

// Implements the `echo.Renderer` interface
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// A View consists of a name and a series of filenames for html components such as base, components and view as well as a template.FuncMap
type View struct {
	Name       string // Name of the view eg: index.html
	Base       []string
	Components []string
	View       string
	Funcs      template.FuncMap
}

// this function provides a mechanism for rendering both full templates and partial templates based on their names.
// The PartialIdentifier constant allows for flexible identification of partial templates, while the apply variable ensures that the correct target template is used for rendering.
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// check if valid template name
	_, ok := t.templates[name]
	if !ok {
		return fmt.Errorf("Error, no template found for: %s\n", name)
	}

	// eg: partial-table.html
	isPartial := strings.HasPrefix(name, PartialIdentifier)
	if isPartial {
		partialRenderTarget := strings.TrimPrefix(name, PartialIdentifier)
		return t.templates[name].ExecuteTemplate(w, partialRenderTarget, data)
	}

	return t.templates[name].ExecuteTemplate(w, DefaultRenderTarget, data)
}

func NewTemplateRegistry(fs fs.FS, templateDir string) (*TemplateRegistry, error) {
	views := getDefinedViews(templateDir)

	templateMap := make(map[string]*template.Template)

	for _, v := range views {
		name := v.Name
		tpl, err := template.New(name).Funcs(v.Funcs).ParseFS(fs, v.Filenames()...)
		if err != nil {
			return nil, err
		}
		templateMap[name] = tpl
	}

	return &TemplateRegistry{
		templates: templateMap,
	}, nil
}

func getBaseLayout(templateDir string) []string {
	return []string{
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, "head.html"),
		filepath.Join(templateDir, "layout.html"),
		filepath.Join(templateDir, "components/header.html"),
	}
}

func getDefinedViews(templateDir string) []View {
	base := getBaseLayout(templateDir)

	return []View{
		View{
			Name: "index.html",
			Base: base,
			View: filepath.Join(templateDir, "views/index.html"),
			Components: []string{
				filepath.Join(templateDir, "components/upload-modal.html"),
			},
		},
		View{
			Name: "upload.html",
			Base: base,
			View: filepath.Join(templateDir, "views/upload.html"),
		},
		View{
			Name: "upload-confirm.html",
			Base: base,
			View: filepath.Join(templateDir, "views/upload-confirm.html"),
		},
		View{
			Name: "login.html",
			Base: base,
			View: filepath.Join(templateDir, "views/login.html"),
		},
		View{
			Name: "account.html",
			Base: base,
			View: filepath.Join(templateDir, "views/account.html"),
		},
		View{
			Name: "cover-letter.html",
			Base: base,
			View: filepath.Join(templateDir, "views/cover-letter.html"),
			Components: []string{
				filepath.Join(templateDir, "components/cover-letter.html"),
			},
		},
		View{
			Name: "cover-letter-print.html",
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
}

// Returns the filenames of all the base files, components and the view filename
func (v *View) Filenames() []string {
	var filenames []string
	filenames = append(filenames, v.Base...)
	filenames = append(filenames, v.Components...)
	filenames = append(filenames, v.View)
	return filenames
}
