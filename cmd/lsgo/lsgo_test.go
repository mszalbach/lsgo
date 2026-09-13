package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/mszalbach/lsgo/internal/explorer"
	"github.com/mszalbach/lsgo/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	root, err := explorer.NewRoot("testdata")
	require.NoError(t, err)
	t.Cleanup(func() {
		err := root.Close()
		if err != nil {
			t.Errorf("failed to clean up resource: %v", err)
		}
	})
	webServer, err := web.NewServer(root)
	require.NoError(t, err)

	testServer := httptest.NewTestServer(t, webServer.Router())
	return testServer
}

func Test_should_serve_static_files(t *testing.T) {
	testCases := map[string]struct {
		url string
	}{
		"SOURCES.md": {url: "http://localhost/static/SOURCES.md"},
		"css":        {url: "http://localhost/static/css/ls.css"},
	}
	// Given
	server := createTestServer(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			res, err := server.Client().Get(tc.url)
			require.NoError(t, err)

			// Then
			assert.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}

func Test_should_list_files_in_directory(t *testing.T) {
	type child struct {
		name string //nolint:unused // checked by the ElementsMatch assert
		href string //nolint:unused // checked by the ElementsMatch assert
		// TODO to check
		// isDir bool   //nolint:unused // checked by the ElementsMatch assert
	}
	testCases := map[string]struct {
		url              string
		expectedChildren []child
	}{
		"root": {
			url: "http://localhost/files",
			expectedChildren: []child{
				{name: "a.md", href: "/files/a.md"},
				{name: "level1", href: "/files/level1"},
			},
		},
		"trailing slash should behave like without trailing slash": {
			url: "http://localhost/files/",
			expectedChildren: []child{
				{name: "a.md", href: "/files/a.md"},
				{name: "level1", href: "/files/level1"},
			},
		},
		"level2": {
			url: "http://localhost/files/level1/level2",
			expectedChildren: []child{
				// TODO the / should not be escaped
				{name: "emptyDir", href: "/files/level1%2Flevel2%2FemptyDir"},
			},
		},
		"empty directory": {
			url:              "http://localhost/files/level1/level2/emptyDir",
			expectedChildren: []child{},
		},
		"links must be correctly encoded or the user could not navigate": {
			url: "http://localhost/files/level1/specialFiles",
			expectedChildren: []child{
				{name: "folder?query=2", href: "/files/level1%2FspecialFiles%2Ffolder%3Fquery=2"},
				{name: "folder#fragment", href: "/files/level1%2FspecialFiles%2Ffolder%23fragment"},
				{
					name: "<a href=\"google.com\">Link file",
					href: "/files/level1%2FspecialFiles%2F%3Ca%20href=%22google.com%22%3ELink%20file",
				},
				{name: "javascript.html", href: "/files/level1%2FspecialFiles%2Fjavascript.html"},
			},
		},
	}
	// Given
	server := createTestServer(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			res, err := server.Client().Get(tc.url)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			// Then
			var actualChildren []child
			doc, err := goquery.NewDocumentFromReader(res.Body)
			require.NoError(t, err)

			doc.Find("td > a:not([href$='?download=1'])").Each(func(_ int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				name := strings.TrimSpace(s.Text())
				actualChildren = append(actualChildren, child{
					name: name,
					href: href,
				})
			})
			assert.ElementsMatch(t, actualChildren, tc.expectedChildren)
		})
	}
}

func Test_should_have_a_breadcrumb_navigation(t *testing.T) {
	type breadcrumb struct {
		name string //nolint:unused // checked by the ElementsMatch assert
		href string //nolint:unused // checked by the ElementsMatch assert
	}

	// Given
	server := createTestServer(t)

	testCases := map[string]struct {
		url                string
		expectedBreadcrump []breadcrumb
	}{
		"root": {url: "http://localhost/files", expectedBreadcrump: []breadcrumb{{name: "Home", href: "/files"}}},
		"level1": {
			url: "http://localhost/files/level1",
			expectedBreadcrump: []breadcrumb{
				{name: "Home", href: "/files"},
				{name: "level1", href: "/files/level1%2F"},
			},
		},
		"links must be correctly encoded or the user could not navigate": {
			url: "http://localhost/files/level1/specialFiles/folder%3Fquery=2",
			expectedBreadcrump: []breadcrumb{
				{name: "Home", href: "/files"},
				{name: "level1", href: "/files/level1%2F"},
				{name: "specialFiles", href: "/files/level1%2FspecialFiles%2F"},
				{name: "folder?query=2", href: "/files/level1%2FspecialFiles%2Ffolder%3Fquery=2%2F"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			res, err := server.Client().Get(tc.url)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			// Then
			doc, err := goquery.NewDocumentFromReader(res.Body)
			require.NoError(t, err)

			var actualBreadcrump []breadcrumb
			doc.Find("nav li > a").Each(func(_ int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				name := strings.TrimSpace(s.Text())
				actualBreadcrump = append(actualBreadcrump, breadcrumb{
					name: name,
					href: href,
				})
			})

			assert.ElementsMatch(t, actualBreadcrump, tc.expectedBreadcrump)
		})
	}
}

func Test_should_serve_files(t *testing.T) {
	// Given
	server := createTestServer(t)

	// When
	res, err := server.Client().Get("http://localhost/files/a.md")
	require.NoError(t, err)

	// Then
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.EqualValues(t, 5, res.ContentLength)
	assert.Equal(t, "text/markdown; charset=utf-8", res.Header.Get("Content-Type"))
	assert.Empty(t, res.Header.Get("Content-Disposition"))
}

func Test_should_provide_file_download(t *testing.T) {
	// Given
	server := createTestServer(t)

	// When
	res, err := server.Client().Get("http://localhost/files/a.md?download=1")
	require.NoError(t, err)

	// Then
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.EqualValues(t, 5, res.ContentLength)
	assert.Equal(t, "text/markdown; charset=utf-8", res.Header.Get("Content-Type"))
	assert.Equal(t, "attachment; filename=a.md", res.Header.Get("Content-Disposition"))
}

func Test_should_have_a_favicon(t *testing.T) {
	// Given
	server := createTestServer(t)

	// When
	res, err := server.Client().Get("http://localhost/favicon.ico")
	require.NoError(t, err)

	// Then
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "image/svg+xml", res.Header.Get("Content-Type"))
}

func Test_should_return_not_found_for_non_existing_resource(t *testing.T) {
	testCases := map[string]struct {
		url string
	}{
		"non existing asset": {url: "http://localhost/static/css/DOES-NOT-EXIST.css"},
		"non existing file":  {url: "http://localhost/files/DOES-NOT-EXIST.md"},
	}
	// Given
	server := createTestServer(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			res, err := server.Client().Get(tc.url)
			require.NoError(t, err)

			// Then
			assert.Equal(t, http.StatusNotFound, res.StatusCode)
		})
	}
}
