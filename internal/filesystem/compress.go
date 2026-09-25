package filesystem

import (
	"archive/zip"
	"fmt"
	"io"
)

// WriteZipArchive creates a zip from the given files and writes it to the io.Writer
func WriteZipArchive(w io.Writer, files ...*File) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()
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
	return nil
}

func addFolder(zipWriter *zip.Writer, folder *File) error {
	subRoot, err := folder.root.osRoot.OpenRoot(folder.RelPath)
	if err != nil {
		return fmt.Errorf("could not open folder %s as fs: %w", folder.RelPath, err)
	}
	defer subRoot.Close()

	err = zipWriter.AddFS(subRoot.FS())
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

	header.Name = file.Name
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
