package ui

import (
	"html/template"
	"io/fs"
	"net/http"
)

type TemplateRenderer struct {
	tpl *template.Template
}

func NewTemplateRenderer(fsys fs.FS) (*TemplateRenderer, error) {
	tpl, err := template.ParseFS(fsys, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &TemplateRenderer{tpl: tpl}, nil
}

func (r *TemplateRenderer) Render(w http.ResponseWriter, name string, data map[string]any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return r.tpl.ExecuteTemplate(w, name, data)
}
