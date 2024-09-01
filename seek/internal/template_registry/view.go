package template_registry

import (
	"html/template"
	"io/fs"
)

type View struct {
	Name       string
	Base       []string
	Components []string
	View       string
	Funcs      template.FuncMap
}

func (v View) Parse(fs fs.FS) (*template.Template, error) {
	var ts []string

	ts = append(ts, v.Base...)
	ts = append(ts, v.Components...)
	ts = append(ts, v.View)

	t, err := template.New(v.Name).Funcs(v.Funcs).ParseFS(fs, ts...)
	if err != nil {
		return nil, err
	}

	return t, nil
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
