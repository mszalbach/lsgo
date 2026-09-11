package web

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/mszalbach/lsgo/internal/assets"
)

type htmlRenderer struct {
	template *template.Template
}

func newHTMLRenderer() (*htmlRenderer, error) {
	sharedTemplates, err := template.New("").ParseFS(assets.Templates, "**/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("could not parse embedded templates: %w", err)
	}

	r := &htmlRenderer{
		template: sharedTemplates,
	}

	return r, nil
}

func (h *htmlRenderer) render(w http.ResponseWriter, status int, data any, templateName string) error {
	var buf bytes.Buffer
	err := h.template.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return fmt.Errorf("could not execute template %s: %w", templateName, err)
	}

	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("could not write template to http response %s: %w", templateName, err)
	}

	return nil
}
