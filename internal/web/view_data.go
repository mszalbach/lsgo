package web

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/mszalbach/lsgo/internal/filesystem"
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

func fileDataFrom(file *filesystem.File) fileData {
	return fileData{
		Name:    file.Name,
		RelPath: relPath{path: file.RelPath},
		ModTime: file.ModTime,
		IsDir:   file.IsDir,
		Size:    file.Size,
	}
}

func folderDataFrom(dir *filesystem.File) ([]fileData, error) {
	children, err := dir.Children()
	if err != nil {
		return nil, fmt.Errorf("could not convert %s to template data: %w", dir.RelPath, err)
	}
	data := make([]fileData, len(children))
	for i, child := range children {
		data[i] = fileDataFrom(&child)
	}

	return data, nil
}

func createBreadcrumb(folderRelPath string) []breadcrumb {
	parts := strings.Split(folderRelPath, "/")
	var breadcrumbs []breadcrumb
	breadcrumbs = append(breadcrumbs, breadcrumb{Name: "Home", RelPath: relPath{path: ""}})
	current := ""
	for _, part := range parts {
		if part != "" && part != "." {
			current = path.Join(current, part)
			breadcrumbs = append(breadcrumbs, breadcrumb{Name: part, RelPath: relPath{path: current}})
		}
	}
	return breadcrumbs
}
