package web

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/mszalbach/lsgo/internal/filesystem"
)

// data contains the overall data structure passed to all templates.
type data struct {
	BaseURL    string
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

// parentBreadcrumb returns the parent folder of the current file/folder. Helper to be used in templates.
func parentBreadcrumb(breadcrumbs []breadcrumb) breadcrumb {
	if len(breadcrumbs) > 1 {
		return breadcrumbs[len(breadcrumbs)-2]
	}
	// Fallback to Home
	return breadcrumbs[0]
}

// currentBreadcrumb returns the current file/folder. Helper to be used in templates.
func currentBreadcrumb(breadcrumbs []breadcrumb) breadcrumb {
	return breadcrumbs[len(breadcrumbs)-1]
}

var sizesSI = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}

const baseSI int64 = 1000

// humanReadableBytes converts a size in bytes to a human-readable string with SI units. Helper to be used in templates.
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

// relPath wraps a relative path string and provides methods for using it in URLs.
type relPath struct {
	path string
}

// String returns the unescaped path.
func (p relPath) String() string {
	return p.path
}

// URLPath returns the path with each segment escaped for use in URLs, preserving path separators.
func (p relPath) URLPath() string {
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

// loadFolderData returns every child that could be read together with any read error.
// Callers may use the returned data even when an error is present.
func loadFolderData(dir *filesystem.File) ([]fileData, error) {
	children, err := dir.Children()
	data := make([]fileData, len(children))
	for i, child := range children {
		data[i] = fileDataFrom(child)
	}

	if err != nil {
		return data, fmt.Errorf("load content of folder %q: %w", dir.RelPath, err)
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
