package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

// HTMLRenderer struct to render go templates for the htlm representation.
type HTMLRenderer struct {
	templateFS fs.FS
	templates  *template.Template
	baseURL    string
}

var funcs = template.FuncMap{
	"bytes": humanReadableBytes,
	"back":  backBreadcrumb,
	"last":  lastBreadcrumb,
}

func backBreadcrumb(breadcrumbs []breadcrumb) breadcrumb {
	if len(breadcrumbs) > 1 {
		return breadcrumbs[len(breadcrumbs)-2]
	}
	// Fallback to most of the time Home
	return breadcrumbs[len(breadcrumbs)-1]
}

func lastBreadcrumb(breadcrumbs []breadcrumb) breadcrumb {
	return breadcrumbs[len(breadcrumbs)-1]
}

// NewHTMLRenderer creates a HtmlRenderer.
// Copied from https://www.alexedwards.net/blog/how-i-use-htmx-with-go
func NewHTMLRenderer(baseURL string, templateFS fs.FS, sharedTemplateFiles ...string) (*HTMLRenderer, error) {
	sharedTemplates, err := template.New("").Funcs(funcs).ParseFS(templateFS, sharedTemplateFiles...)
	if err != nil {
		return nil, fmt.Errorf("could not parse embedded templates: %w", err)
	}

	r := &HTMLRenderer{
		templateFS: templateFS,
		templates:  sharedTemplates,
		baseURL:    baseURL,
	}

	return r, nil
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
		return fmt.Errorf("could not clone templates %s: %w", templateName, err)
	}

	if len(additionalTemplateFiles) > 0 {
		ts, err = ts.ParseFS(h.templateFS, additionalTemplateFiles...)
		if err != nil {
			return fmt.Errorf("could not parse additional templates %v: %w", additionalTemplateFiles, err)
		}
	}

	buf := new(bytes.Buffer)
	data.BaseURL = h.baseURL
	err = ts.ExecuteTemplate(buf, templateName, data)
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

var sizesSI = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}

const baseSI int64 = 1000

func humanReadableBytes(size int64) string {
	if size < 0 {
		return "0 B"
	}

	unitsLimit := len(sizesSI) - 1
	i := 0

	// Keep dividing until size is under 1000 or we hit the maximum unit (EB)
	for size >= baseSI && i < unitsLimit {
		size /= baseSI
		i++
	}

	return fmt.Sprintf("%d %s", size, sizesSI[i])
}
