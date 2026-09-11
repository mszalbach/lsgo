// Package explorer contains everything to work with a directory and it childs
package explorer

import (
	"os"
	"path"
	"time"
)

// Root represents a folder structure
type Root struct {
	root *os.Root
}

// NewRoot creates a Root to work with a folder structure
func NewRoot(name string) (Root, error) {
	root, err := os.OpenRoot(name)
	if err != nil {
		return Root{}, err
	}

	return Root{
		root: root,
	}, nil
}

// File represents a folder and file in the Root.
// This simplifies the handling with files because the important information are provided in one struct, instead of os.File.
type File struct {
	root    *os.Root
	relPath string
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// AsOsFile is used when the underlying os.File is needed.
// Ensure to close it after usage.
func (f *File) AsOsFile() (*os.File, error) {
	return f.root.Open(f.relPath)
}

// Children return the children of the current File. Will return nil if the file is not a directory.
func (f *File) Children() ([]File, error) {
	if !f.IsDir {
		return nil, nil
	}

	dir, err := f.root.Open(f.relPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	dirEntries, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	children := make([]File, len(dirEntries))
	for i, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		children[i] = File{
			root:    f.root,
			relPath: path.Join(f.relPath, entry.Name()),
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
	}

	return children, nil
}

// File returns a File from the current Root.
func (r *Root) File(name string) (*File, error) {
	rootFile, err := r.root.Open(name)
	if err != nil {
		return nil, err
	}
	defer rootFile.Close()
	stat, err := rootFile.Stat()
	if err != nil {
		return nil, err
	}

	return &File{
		Name:    stat.Name(),
		IsDir:   stat.IsDir(),
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
		root:    r.root,
		relPath: name,
	}, nil
}
