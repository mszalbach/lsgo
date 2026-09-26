package filesystem

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
)

// WriteZipArchive creates a zip from the given files and writes it to the io.Writer.
func WriteZipArchive(w io.Writer, files ...*File) error {
	zipWriter := zip.NewWriter(w)

	for _, file := range files {
		if file.IsDir {
			err := addFolder(zipWriter, file)
			if err != nil {
				return err
			}
		} else {
			err := addFile(zipWriter, file)
			if err != nil {
				return err
			}
		}
	}

	err := zipWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to finalize zip archive: %w", err)
	}

	return nil
}

func addFolder(zipWriter *zip.Writer, folder *File) error {
	subFS, err := folder.asFS()
	if err != nil {
		return fmt.Errorf("failed to get sub fs %s for adding to zip: %w", folder.RelPath, err)
	}

	err = addFS(zipWriter, folder.RelPath, subFS)
	if err != nil {
		return fmt.Errorf("failed to add folder %s to zip: %w", folder.RelPath, err)
	}
	return nil
}

func addFile(zipWriter *zip.Writer, file *File) error {
	osFile, err := file.AsOSFile()
	if err != nil {
		return fmt.Errorf("failed to add file %s to zip: %w", file.RelPath, err)
	}
	defer osFile.Close()

	osStats, err := osFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats %s: %w", file.RelPath, err)
	}

	header, err := zip.FileInfoHeader(osStats)
	if err != nil {
		return fmt.Errorf("failed to create file info header for file %s to zip: %w", file.RelPath, err)
	}

	header.Name = file.RelPath
	header.Method = zip.Deflate

	fileHeader, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %s: %w", file.RelPath, err)
	}

	_, err = io.Copy(fileHeader, osFile)
	if err != nil {
		return fmt.Errorf("failed to add file %s to zip: %w", file.RelPath, err)
	}
	return nil
}

func addFS(w *zip.Writer, basePath string, fsys fs.FS) error {
	err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		return addFSEntry(w, basePath, fsys, name, d)
	})
	if err != nil {
		return fmt.Errorf("failed to add fs %s to zip: %w", basePath, err)
	}
	return nil
}

func addFSEntry(w *zip.Writer, basePath string, fsys fs.FS, name string, d fs.DirEntry) error {
	zipPath := fsZipPath(basePath, name)
	if zipPath == "" {
		return nil
	}

	info, err := d.Info()
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", name, err)
	}
	if !d.IsDir() && !info.Mode().IsRegular() {
		return fmt.Errorf("failed to add non-regular file %s", name)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %s: %w", name, err)
	}
	header.Name = zipPath
	if d.IsDir() {
		header.Name += "/"
	}
	header.Method = zip.Deflate

	entryWriter, err := w.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to add %s to zip: %w", name, err)
	}
	if d.IsDir() {
		return nil
	}

	return copyFSFile(fsys, name, entryWriter)
}

func fsZipPath(basePath string, name string) string {
	if name == "." {
		return basePath
	}
	if basePath == "" {
		return name
	}
	return basePath + "/" + name
}

func copyFSFile(fsys fs.FS, name string, dst io.Writer) error {
	f, err := fsys.Open(name)
	if err != nil {
		return fmt.Errorf("failed to open file %s to add to zip: %w", name, err)
	}
	defer f.Close()

	_, err = io.Copy(dst, f)
	if err != nil {
		return fmt.Errorf("failed to write file %s to zip: %w", name, err)
	}
	return nil
}
