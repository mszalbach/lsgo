package web

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/mszalbach/lsgo/internal/explorer"
)

// data the overal struct given to all templates.
type data struct {
	Breadcrumb []breadcrumb
	Content    any
}

// fileData contains the information needed to render a file in HTML.
type fileData struct {
	RelPath relPath
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// breadcrumb contains the parent folder names and paths used to construct links.
type breadcrumb struct {
	Name    string
	RelPath relPath
}

// relPath ensures that paths are properly escaped while preserving path separators.
type relPath struct {
	path string
}

// String returns the encoded path while preserving the "/" separators.
func (p relPath) String() string {
	parts := strings.Split(p.path, "/")
	var escapedURL []string

	for _, part := range parts {
		if part != "" {
			escapedURL = append(escapedURL, url.PathEscape(part))
		}
	}

	return path.Join(escapedURL...)
}

func fileDataFrom(file *explorer.File) fileData {
	return fileData{
		Name:    file.Name,
		RelPath: relPath{path: file.RelPath},
		ModTime: file.ModTime,
		IsDir:   file.IsDir,
		Size:    file.Size,
	}
}

func directoryDataFrom(dir *explorer.File) ([]fileData, error) {
	children, err := dir.Children()
	if err != nil {
		return nil, fmt.Errorf("could not convert %s to template data: %w", dir.RelPath, err)
	}
	data := make([]fileData, 0, len(children))

	for _, child := range children {
		data = append(data, fileDataFrom(&child))
	}

	return data, nil
}

func createBreadcrumb(directoryRelPath string) []breadcrumb {
	parts := strings.Split(directoryRelPath, "/")
	var breadcrumbs []breadcrumb
	current := ""
	for _, part := range parts {
		if part != "" && part != "." {
			current = path.Join(current, part)
			breadcrumbs = append(breadcrumbs, breadcrumb{Name: part, RelPath: relPath{path: current}})
		}
	}
	return breadcrumbs
}
