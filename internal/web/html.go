package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

// HTMLRenderer struct to render go templates for the html representation.
type HTMLRenderer struct {
	templateFS fs.FS
	templates  *template.Template
	baseURL    string
}

var funcs = template.FuncMap{
	"bytes":             humanReadableBytes,
	"parentBreadcrumb":  parentBreadcrumb,
	"currentBreadcrumb": currentBreadcrumb,
}

// NewHTMLRenderer creates a HTMLRenderer.
// Copied from https://www.alexedwards.net/blog/how-i-use-htmx-with-go
func NewHTMLRenderer(baseURL string, templateFS fs.FS, sharedTemplateFiles ...string) (*HTMLRenderer, error) {
	sharedTemplates, err := template.New("").Funcs(funcs).ParseFS(templateFS, sharedTemplateFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded templates: %w", err)
	}

	return &HTMLRenderer{
		templateFS: templateFS,
		templates:  sharedTemplates,
		baseURL:    baseURL,
	}, nil
}

func (h *HTMLRenderer) render(
	w http.ResponseWriter,
	status int,
	data data,
	templateName string,
	additionalTemplateFiles ...string,
) error {
	ts, err := h.templates.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone templates %s: %w", templateName, err)
	}

	if len(additionalTemplateFiles) > 0 {
		ts, err = ts.ParseFS(h.templateFS, additionalTemplateFiles...)
		if err != nil {
			return fmt.Errorf("failed to parse additional templates %v: %w", additionalTemplateFiles, err)
		}
	}

	buf := new(bytes.Buffer)
	data.BaseURL = h.baseURL
	err = ts.ExecuteTemplate(buf, templateName, data)
	if err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	if err != nil {
		return fmt.Errorf("failed to write template to http response %s: %w", templateName, err)
	}

	return nil
}
