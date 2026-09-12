package web

import (
	"fmt"
	"net/url"
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
	RelPath string
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// breadcrumb navigation to have the names of all parent folders and there path to construct links
type breadcrumb struct {
	Name    string
	RelPath string
}

func fileDataFrom(file *explorer.File) fileData {
	return fileData{
		Name:    file.Name,
		RelPath: url.PathEscape(file.RelPath),
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
