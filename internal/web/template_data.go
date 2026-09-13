package web

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/mszalbach/lsgo/internal/explorer"
)

// directoryData presentation of a directory for the web ui
type directoryData struct {
	Breadcrumb []breadcrumb
	Children   []fileData
}

// fileData file information needed to render it as html
type fileData struct {
	RelPath relPath
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// breadcrumb navigation to have the names of all parent folders and there path to construct links
type breadcrumb struct {
	Name    string
	RelPath relPath
}

// relPath helper to ensure paths are correctly path escaped in all template data
type relPath struct {
	path string
}

// String returns the string representation which ensures url special characters are encoded by still keeping the "/" segements
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

func directoryDataFrom(dir *explorer.File) (directoryData, error) {
	children, err := dir.Children()
	if err != nil {
		return directoryData{}, fmt.Errorf("could not convert %s to template data: %w", dir.RelPath, err)
	}

	data := directoryData{
		Breadcrumb: createBreadcrumb(dir.RelPath),
		Children:   make([]fileData, 0, len(children)),
	}
	for _, child := range children {
		data.Children = append(data.Children, fileDataFrom(&child))
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
