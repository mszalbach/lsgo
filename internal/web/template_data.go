package web

import (
	"fmt"
	"net/url"
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

func (p relPath) String() string {
	return url.PathEscape(p.path)
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

func createBreadcrumb(path string) []breadcrumb {
	parts := strings.Split(path, "/")
	var breadcrumbs []breadcrumb
	current := ""
	for _, part := range parts {
		if part != "" && part != "." {
			current = current + part + "/"
			breadcrumbs = append(breadcrumbs, breadcrumb{Name: part, RelPath: relPath{path: current}})
		}
	}
	return breadcrumbs
}
