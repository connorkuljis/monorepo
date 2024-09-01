package template_registry

import (
	"html/template"
	"io/fs"
)

type Partial struct {
	Name       string
	Components []string
}

func (p Partial) Parse(fs fs.FS) (*template.Template, error) {
	t, err := template.ParseFS(fs, p.Components...)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func LoadPartials(templates map[string]*template.Template, partials []Partial, fs fs.FS) (map[string]*template.Template, error) {
	for _, p := range partials {
		t, err := p.Parse(fs)
		if err != nil {
			return nil, err
		}
		templates[p.Name] = t
	}
	return templates, nil
}
