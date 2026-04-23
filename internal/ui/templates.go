package ui

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
}

// NewTemplateRenderer parseia cada par (layout + page) em um template.Template separado,
// evitando conflito entre múltiplos `{{define "content"}}` quando parseados no mesmo set.
func NewTemplateRenderer(fsys fs.FS) (*TemplateRenderer, error) {
	pages := []string{"login", "app"}
	out := make(map[string]*template.Template, len(pages))
	for _, p := range pages {
		tpl, err := template.ParseFS(fsys, "templates/layout.html", "templates/"+p+".html")
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		out[p] = tpl
	}
	return &TemplateRenderer{templates: out}, nil
}

func (r *TemplateRenderer) Render(w http.ResponseWriter, name string, data map[string]any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template não encontrado: %s", name)
	}
	return tpl.ExecuteTemplate(w, name, data)
}
