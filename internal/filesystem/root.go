// Package filesystem contains everything needed to work with a folder and its children.
package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"
)

// ErrPartialChildren indicates that some directory entries could not be read and the response can be used for partial listing.
var ErrPartialChildren = errors.New("some directory entries could not be read")

// Root represents a folder and its contents.
type Root struct {
	osRoot *os.Root
}

// NewRoot creates a Root for working with a folder and its contents.
func NewRoot(rootPath string) (Root, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return Root{}, fmt.Errorf("failed to create root %s: %w", rootPath, err)
	}

	return Root{
		osRoot: root,
	}, nil
}

// File represents either a file or a directory within a Root.
// This simplifies file handling because the important information is kept in one struct instead of an os.File.
type File struct {
	root    *Root
	RelPath string
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

// AsOSFile is used when the underlying os.File is needed.
// Make sure to close it after use.
func (f *File) AsOSFile() (*os.File, error) {
	file, err := f.root.osRoot.Open(f.RelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", f.RelPath, err)
	}
	return file, nil
}

// asFS will return a sub filesystem from the root.
// So the relative folder paths are kept, which is needed for correctly zipping.
func (f *File) asFS() (fs.FS, error) {
	subFS, err := fs.Sub(f.root.osRoot.FS(), f.RelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub fs for %s: %w", f.RelPath, err)
	}

	return subFS, nil
}

// Children returns the children of the current File. It returns nil if the File is not a directory.
// If an entry cannot be read, successfully read children are returned with the errors joined.
// So handle partial errors.
func (f *File) Children() ([]*File, error) {
	if !f.IsDir {
		return nil, nil
	}

	osDir, err := f.AsOSFile()
	if err != nil {
		return nil, fmt.Errorf("failed to use os file %s to get children: %w", f.RelPath, err)
	}
	defer osDir.Close()

	dirEntries, err := osDir.ReadDir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read folder %s: %w", f.RelPath, err)
	}

	var readErrors []error
	var children []*File
	for _, entry := range dirEntries {
		// Entries stats would not follow symlinks and the isDir property would be wrong so correctly open the child
		file, err := f.root.File(path.Join(f.RelPath, entry.Name()))
		if err != nil {
			readErrors = append(
				readErrors,
				fmt.Errorf("failed to read child %s in folder %s: %w", entry.Name(), f.RelPath, err),
			)
			continue
		}
		children = append(children, file)
	}

	if len(readErrors) > 0 {
		readErrors = append(readErrors, ErrPartialChildren)
	}
	return children, errors.Join(readErrors...)
}

// File returns a File from the current Root.
func (r *Root) File(filePath string) (*File, error) {
	stat, err := r.osRoot.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for %s: %w", filePath, err)
	}

	return &File{
		Name:    stat.Name(),
		IsDir:   stat.IsDir(),
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
		root:    r,
		RelPath: filePath,
	}, nil
}

// Close closes the Root.
func (r *Root) Close() error {
	err := r.osRoot.Close()
	if err != nil {
		return fmt.Errorf("failed to close root: %w", err)
	}
	return nil
}
