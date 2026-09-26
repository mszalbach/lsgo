package filesystem

import (
	"archive/zip"
	"errors"
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
				return fmt.Errorf("could not add folder %s to zip: %w", file.RelPath, err)
			}
		} else {
			err := addFile(zipWriter, file)
			if err != nil {
				return fmt.Errorf("could not add file %s to zip: %w", file.RelPath, err)
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
	subFS, err := folder.AsFS()
	if err != nil {
		return fmt.Errorf("could not get sub fs %s for adding to zip: %w", folder.RelPath, err)
	}

	err = addFS(zipWriter, folder.RelPath, subFS)
	if err != nil {
		return fmt.Errorf("could not add folder %s to zip: %w", folder.RelPath, err)
	}
	return nil
}

func addFile(zipWriter *zip.Writer, file *File) error {
	osFile, err := file.AsOsFile()
	if err != nil {
		return fmt.Errorf("could not add file %s to zip: %w", file.RelPath, err)
	}
	defer osFile.Close()

	osStats, err := osFile.Stat()
	if err != nil {
		return fmt.Errorf("could not get file stats %s: %w", file.RelPath, err)
	}

	header, err := zip.FileInfoHeader(osStats)
	if err != nil {
		return fmt.Errorf("could not create file info header for file %s to zip: %w", file.RelPath, err)
	}

	header.Name = file.RelPath
	header.Method = zip.Deflate

	fileHeader, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("could not create zip header %s to zip: %w", file.RelPath, err)
	}

	_, err = io.Copy(fileHeader, osFile)
	if err != nil {
		return fmt.Errorf("could not add file %s to zip: %w", file.RelPath, err)
	}
	return nil
}

//nolint:cyclop // blabla
func addFS(w *zip.Writer, basePath string, fsys fs.FS) error {
	err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate zip internal path
		var zipPath string
		switch {
		case name == ".":
			// root folder must be added as basePath or it would be missing
			zipPath = basePath
		case basePath == "":
			zipPath = name
		default:
			zipPath = basePath + "/" + name
		}

		// Skip empty root directory entries
		if zipPath == "" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("could not get file info while adding fs %s to zip: %w", basePath, err)
		}
		if !d.IsDir() && !info.Mode().IsRegular() {
			return errors.New("cannot add non-regular file")
		}

		h, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("could not create file for %s to zip: %w", name, err)
		}

		h.Name = zipPath
		if d.IsDir() {
			h.Name += "/"
		}
		h.Method = zip.Deflate

		fw, err := w.CreateHeader(h)
		if err != nil {
			return fmt.Errorf("could not add file %s to zip: %w", name, err)
		}

		if d.IsDir() {
			return nil
		}

		f, err := fsys.Open(name)
		if err != nil {
			return fmt.Errorf("could not open file %s to add to zip: %w", name, err)
		}
		defer f.Close()

		_, err = io.Copy(fw, f)
		if err != nil {
			return fmt.Errorf("could not write file %s to zip: %w", name, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("could not add fs %s to zip: %w", basePath, err)
	}
	return nil
}
