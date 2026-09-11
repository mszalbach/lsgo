package web

import (
	"bytes"
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
		return nil, err
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
		return err
	}

	w.WriteHeader(status)
	buf.WriteTo(w)

	return nil
}
