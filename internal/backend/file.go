package backend

import (
	"os"
	"path"
	"time"
)

type rootFS struct {
	root *os.Root
}

func NewRootFS(name string) (*rootFS, error) {
	root, err := os.OpenRoot(name)
	if err != nil {
		return nil, err
	}

	return &rootFS{
		root: root,
	}, nil
}

type file struct {
	root    *os.Root
	relPath string
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

func (f file) AsOsFile() (*os.File, error) {
	return f.root.Open(f.relPath)
}

func (f file) IsRoot() bool {
	return f.Name == "."

}

func (f file) Children() ([]file, error) {
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

	children := make([]file, len(dirEntries))
	for i, entry := range dirEntries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		children[i] = file{
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

func (r *rootFS) File(name string) (file, error) {
	rootFile, err := r.root.Open(name)
	if err != nil {
		return file{}, err
	}
	defer rootFile.Close()
	stat, err := rootFile.Stat()
	if err != nil {
		return file{}, err
	}

	file := file{
		Name:    stat.Name(),
		IsDir:   stat.IsDir(),
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
		root:    r.root,
		relPath: name,
	}

	return file, nil
}
