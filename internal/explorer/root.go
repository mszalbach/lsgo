package explorer

import (
	"os"
	"path"
	"time"
)

type Root struct {
	root *os.Root
}

func NewRoot(name string) (*Root, error) {
	root, err := os.OpenRoot(name)
	if err != nil {
		return nil, err
	}

	return &Root{
		root: root,
	}, nil
}

type File struct {
	root    *os.Root
	relPath string
	Name    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

func (f File) AsOsFile() (*os.File, error) {
	return f.root.Open(f.relPath)
}

func (f File) IsRoot() bool {
	return f.Name == "."

}

func (f File) Children() ([]File, error) {
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

func (r *Root) File(name string) (File, error) {
	rootFile, err := r.root.Open(name)
	if err != nil {
		return File{}, err
	}
	defer rootFile.Close()
	stat, err := rootFile.Stat()
	if err != nil {
		return File{}, err
	}

	file := File{
		Name:    stat.Name(),
		IsDir:   stat.IsDir(),
		Size:    stat.Size(),
		ModTime: stat.ModTime(),
		root:    r.root,
		relPath: name,
	}

	return file, nil
}
