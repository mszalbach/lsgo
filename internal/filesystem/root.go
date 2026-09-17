// Package filesystem contains everything needed to work with a directory and its children.
package filesystem

import (
	"fmt"
	"os"
	"path"
	"time"
)

// Root represents a folder structure.
type Root struct {
	root *os.Root
}

// NewRoot creates a Root to work with a folder structure.
func NewRoot(name string) (Root, error) {
	root, err := os.OpenRoot(name)
	if err != nil {
		return Root{}, fmt.Errorf("could not create Root %s: %w", name, err)
	}

	return Root{
		root: root,
	}, nil
}

// File represents a folder or file in the Root.
// This simplifies file handling because the important information is provided in one struct instead of an os.File.
type File struct {
	root    *os.Root
	RelPath string
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// AsOsFile is used when the underlying os.File is needed.
// Ensure to close it after usage.
func (f *File) AsOsFile() (*os.File, error) {
	file, err := f.root.Open(f.RelPath)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", f.RelPath, err)
	}

	return file, nil
}

// Children returns the children of the current File. It returns nil if the file is not a directory.
func (f *File) Children() ([]File, error) {
	if !f.IsDir {
		return nil, nil
	}

	dir, err := f.root.Open(f.RelPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s to get children: %w", f.RelPath, err)
	}
	defer dir.Close()

	dirEntries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("could not read the directory %s: %w", f.RelPath, err)
	}

	var children []File
	for _, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		children = append(children, File{
			root:    f.root,
			RelPath: path.Join(f.RelPath, entry.Name()),
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	return children, nil
}

// File returns a File from the current Root.
func (r *Root) File(name string) (*File, error) {
	rootFile, err := r.root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", name, err)
	}
	defer rootFile.Close()
	stat, err := rootFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not get stats for %s: %w", name, err)
	}

	return &File{
		Name:    stat.Name(),
		IsDir:   stat.IsDir(),
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
		root:    r.root,
		RelPath: name,
	}, nil
}

// Close closes the Root.
func (r *Root) Close() error {
	err := r.root.Close()
	if err != nil {
		return fmt.Errorf("could not close root: %w", err)
	}
	return nil
}
