package filesystem_test

import (
	"testing"

	"github.com/mszalbach/lsgo/internal/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestdataRoot(t *testing.T) filesystem.Root {
	t.Helper()
	root, err := filesystem.NewRoot("testdata")
	require.NoError(t, err)
	t.Cleanup(func() {
		err := root.Close()
		if err != nil {
			t.Errorf("failed to clean up resource: %v", err)
		}
	})
	return root
}

func Test_root_returns_information_about_file(t *testing.T) {
	testCases := map[string]struct {
		name            string
		isDir           bool
		expectedName    string
		expectedRelPath string
	}{
		"./hello.md": {
			name:            "hello.md",
			isDir:           false,
			expectedName:    "hello.md",
			expectedRelPath: "hello.md",
		},
		"./folder/folderInFolder": {
			name:            "folder/folderInFolder",
			isDir:           true,
			expectedName:    "folderInFolder",
			expectedRelPath: "folder/folderInFolder",
		},
		"./folder/folderInFolder/a.yaml": {
			name:            "folder/folderInFolder/a.yaml",
			isDir:           false,
			expectedName:    "a.yaml",
			expectedRelPath: "folder/folderInFolder/a.yaml",
		},
	}

	// Given
	root := createTestdataRoot(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			file, err := root.File(tc.name)
			require.NoError(t, err)

			// Then
			assert.Equal(t, tc.isDir, file.IsDir)
			assert.Equal(t, tc.expectedName, file.Name)
			assert.Equal(t, tc.expectedRelPath, file.RelPath)
		})
	}
}

func Test_file_returns_its_children(t *testing.T) {
	type child struct {
		name    string //nolint:unused // checked by the ElementsMatch assert
		relPath string //nolint:unused // checked by the ElementsMatch assert
	}
	testCases := map[string]struct {
		name             string
		expectedChildren []child
	}{
		"hello.md": {
			name:             "hello.md",
			expectedChildren: []child{},
		},
		"folderInFolder": {
			name: "folder/folderInFolder",
			expectedChildren: []child{
				{name: "a.yaml", relPath: "folder/folderInFolder/a.yaml"},
				{name: "b.yaml", relPath: "folder/folderInFolder/b.yaml"},
			},
		},
	}

	// Given
	root := createTestdataRoot(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			file, err := root.File(tc.name)
			require.NoError(t, err)

			children, err := file.Children()
			require.NoError(t, err)

			var actualChildren []child

			for _, file := range children {
				actualChildren = append(actualChildren, child{name: file.Name, relPath: file.RelPath})
			}

			// Then
			assert.ElementsMatch(t, actualChildren, tc.expectedChildren)
		})
	}
}

func Test_file_can_be_opened_as_an_os_file(t *testing.T) {
	// Given
	root := createTestdataRoot(t)

	file, err := root.File("hello.md")
	require.NoError(t, err)

	// When
	osFile, err := file.AsOsFile()
	require.NoError(t, err)
	t.Cleanup(func() {
		err := osFile.Close()
		if err != nil {
			t.Errorf("failed to clean up resource: %v", err)
		}
	})

	// Then
	assert.Equal(t, "testdata/hello.md", osFile.Name())
}

func Test_root_cannot_be_created_for_a_nonexistent_folder(t *testing.T) {
	testCases := map[string]struct {
		path string
	}{
		"nonexistent folder": {
			path: "testdata/DOES_NOT_EXIST",
		},
		"file": {
			path: "testdata/hello.md",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			// When
			root, err := filesystem.NewRoot(tc.path)

			// Then
			assert.Empty(t, root)
			assert.Error(t, err)
		})
	}
}

func Test_root_cannot_access_files_outside_its_folder(t *testing.T) {
	testCases := map[string]struct {
		name string
	}{
		"path traversal": {
			name: "../root_test.go",
		},
		"absolute file linux": {
			name: "/tmp",
		},
		"absolute file windows": {
			name: "C:\\Users\\Public",
		},
	}

	// Given
	root := createTestdataRoot(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			// When
			root, err := root.File(tc.name)

			// Then
			assert.Empty(t, root)
			assert.Error(t, err)
		})
	}
}
