package filesystem_test

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/mszalbach/lsgo/internal/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_should_have_correct_paths_related_to_root(t *testing.T) {
	// Given
	root := createTestdataRoot(t)

	file, err := root.File("folder")
	require.NoError(t, err)

	// When
	var buf bytes.Buffer
	err = filesystem.WriteZipArchive(&buf, file)
	require.NoError(t, err)

	// Then
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	var actualFileNames []string
	for _, file := range zipReader.File {
		actualFileNames = append(actualFileNames, file.Name)
	}

	// everything must start from folder/ and not treat folder/ as "."
	expectedFileNames := []string{
		"folder/",
		"folder/folderInFolder/",
		"folder/folderInFolder/a.yaml",
		"folder/folderInFolder/b.yaml",
	}

	assert.ElementsMatch(t, expectedFileNames, actualFileNames)
}
