package explorer_test

import (
	"testing"

	"github.com/mszalbach/lsgo/internal/explorer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_root_returns_information_about_file(t *testing.T) {
	testCases := map[string]struct {
		name            string
		isDir           bool
		expectedName    string
		expextedRelPath string
	}{
		"./hello.md": {
			name:            "hello.md",
			isDir:           false,
			expectedName:    "hello.md",
			expextedRelPath: "hello.md",
		},
		"./folder/folderInFolder": {
			name:            "folder/folderInFolder",
			isDir:           true,
			expectedName:    "folderInFolder",
			expextedRelPath: "folder/folderInFolder",
		},
		"./folder/folderInFolder/a.yaml": {
			name:            "folder/folderInFolder/a.yaml",
			isDir:           false,
			expectedName:    "a.yaml",
			expextedRelPath: "folder/folderInFolder/a.yaml",
		},
	}

	// Given
	root, err := explorer.NewRoot("testdata")
	require.NoError(t, err)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			file, err := root.File(tc.name)
			require.NoError(t, err)

			// Then
			assert.Equal(t, tc.isDir, file.IsDir)
			assert.Equal(t, tc.expectedName, file.Name)
			assert.Equal(t, tc.expextedRelPath, file.RelPath)
		})
	}
}

func Test_file_returns_its_children(t *testing.T) {
	testCases := map[string]struct {
		name             string
		expectedChildren []string
	}{
		"hello.md": {
			name:             "hello.md",
			expectedChildren: nil,
		},
		"folderInFolder": {
			name:             "folder/folderInFolder",
			expectedChildren: []string{"a.yaml", "b.yaml"},
		},
	}

	// Given
	root, err := explorer.NewRoot("testdata")
	require.NoError(t, err)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			file, err := root.File(tc.name)
			require.NoError(t, err)

			children, err := file.Children()
			require.NoError(t, err)

			var actualNames []string

			for _, file := range children {
				actualNames = append(actualNames, file.Name)
			}

			// Then
			assert.ElementsMatch(t, actualNames, tc.expectedChildren)
		})
	}
}

func Test_file_can_be_opened_as_os_file_to_use_it_with_other_go_functions(t *testing.T) {
	// Given
	root, err := explorer.NewRoot("testdata")
	require.NoError(t, err)

	file, err := root.File("hello.md")
	require.NoError(t, err)

	// When
	osFile, err := file.AsOsFile()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := osFile.Close(); err != nil {
			t.Errorf("failed to clean up resource: %v", err)
		}
	})

	// Then
	assert.Equal(t, "testdata/hello.md", osFile.Name())
}

func Test_root_can_not_be_created_for_non_existing_folder(t *testing.T) {
	testCases := map[string]struct {
		path string
	}{
		"non existing folder": {
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
			root, err := explorer.NewRoot(tc.path)

			// Then
			assert.Empty(t, root)
			assert.Error(t, err)
		})
	}
}

func Test_root_can_not_access_files_out_of_his_directory_structure(t *testing.T) {
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
	root, err := explorer.NewRoot("testdata")
	require.NoError(t, err)

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
