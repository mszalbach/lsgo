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

var funcs = template.FuncMap{
	"bytes": humanReadableBytes,
}

func newHTMLRenderer() (*htmlRenderer, error) {
	sharedTemplates, err := template.New("").Funcs(funcs).ParseFS(assets.Templates, "**/*.tmpl")
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

var sizesSI = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}

const baseSI int64 = 1000

func humanReadableBytes(size int64) string {
	if size < 0 {
		return "0 B"
	}

	unitsLimit := len(sizesSI) - 1
	i := 0

	// Keep dividing until size is under 1024 or we hit the maximum unit (EB)
	for size >= baseSI && i < unitsLimit {
		size /= baseSI
		i++
	}

	return fmt.Sprintf("%d %s", size, sizesSI[i])
}
