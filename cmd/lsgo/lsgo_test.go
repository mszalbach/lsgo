package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/mszalbach/lsgo/internal/assets"
	"github.com/mszalbach/lsgo/internal/filesystem"
	"github.com/mszalbach/lsgo/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	err := os.MkdirAll("testdata/level1/level2/emptyDir", 0o750)
	if err != nil {
		//nolint:forbidigo // in TestMain there is no default logger
		fmt.Println("Could not create required empty folder")
		os.Exit(1)
	}
	m.Run()
}

func createTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return createTestServerWithPath(t, "/")
}

func createTestServerWithPath(t *testing.T, baseURL string) *httptest.Server {
	t.Helper()
	root, err := filesystem.NewRoot("testdata")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, root.Close()) })
	renderer, err := web.NewHTMLRenderer(baseURL, assets.Templates, "html/base.tmpl")
	require.NoError(t, err)
	webServer, err := web.NewRouter(root, renderer, 5)
	require.NoError(t, err)

	testServer := httptest.NewTestServer(t, webServer.Routes())
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

func Test_should_list_files_in_folder(t *testing.T) {
	type child struct {
		name  string //nolint:unused // checked by the ElementsMatch assert
		href  string //nolint:unused // checked by the ElementsMatch assert
		isDir bool   //nolint:unused // checked by the ElementsMatch assert
	}
	testCases := map[string]struct {
		url              string
		expectedChildren []child
	}{
		"root": {
			url: "http://localhost/files",
			expectedChildren: []child{
				{name: "a.md", href: "files/a.md", isDir: false},
				{name: "level1", href: "files/level1", isDir: true},
			},
		},
		"trailing slash should behave like without trailing slash": {
			url: "http://localhost/files/",
			expectedChildren: []child{
				{name: "a.md", href: "files/a.md", isDir: false},
				{name: "level1", href: "files/level1", isDir: true},
			},
		},
		"level2": {
			url: "http://localhost/files/level1/level2",
			expectedChildren: []child{
				{name: "emptyDir", href: "files/level1/level2/emptyDir", isDir: true},
			},
		},
		"empty folder": {
			url:              "http://localhost/files/level1/level2/emptyDir",
			expectedChildren: []child{},
		},
		"links must be correctly encoded or the user could not navigate": {
			url: "http://localhost/files/level1/specialFiles",
			expectedChildren: []child{
				{name: "folder?query=2", href: "files/level1/specialFiles/folder%3Fquery=2", isDir: true},
				{name: "folder#fragment", href: "files/level1/specialFiles/folder%23fragment", isDir: true},
				{
					name: "<a href=\"google.com\">Link file",
					href: "files/level1/specialFiles/%3Ca%20href=%22google.com%22%3ELink%20file",
				},
				{name: "javascript.html", href: "files/level1/specialFiles/javascript.html", isDir: false},
				{
					name:  "html-without-extension",
					href:  "files/level1/specialFiles/html-without-extension",
					isDir: false,
				},
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
				alt, _ := s.Find("img").Attr("alt")
				isDir := alt == "folder"

				actualChildren = append(actualChildren, child{
					name:  name,
					href:  href,
					isDir: isDir,
				})
			})
			assert.ElementsMatch(t, actualChildren, tc.expectedChildren)
		})
	}
}

func Test_should_have_breadcrumb_navigation(t *testing.T) {
	type breadcrumb struct {
		name string //nolint:unused // checked by the ElementsMatch assert
		href string //nolint:unused // checked by the ElementsMatch assert
	}

	// Given
	server := createTestServer(t)

	testCases := map[string]struct {
		url                string
		expectedBreadcrumb []breadcrumb
	}{
		"root": {url: "http://localhost/files", expectedBreadcrumb: []breadcrumb{{name: "Home", href: "files/"}}},
		"level1": {
			url: "http://localhost/files/level1",
			expectedBreadcrumb: []breadcrumb{
				{name: "Home", href: "files/"},
				{name: "level1", href: "files/level1"},
			},
		},
		"links must be correctly encoded or the user could not navigate": {
			url: "http://localhost/files/level1/specialFiles/folder%3Fquery=2",
			expectedBreadcrumb: []breadcrumb{
				{name: "Home", href: "files/"},
				{name: "level1", href: "files/level1"},
				{name: "specialFiles", href: "files/level1/specialFiles"},
				{name: "folder?query=2", href: "files/level1/specialFiles/folder%3Fquery=2"},
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

			var actualBreadcrumb []breadcrumb
			doc.Find("nav li > a").Each(func(_ int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				name := strings.TrimSpace(s.Text())
				actualBreadcrumb = append(actualBreadcrumb, breadcrumb{
					name: name,
					href: href,
				})
			})

			assert.ElementsMatch(t, actualBreadcrumb, tc.expectedBreadcrumb)
		})
	}
}

func Test_should_serve_files(t *testing.T) {
	testCases := map[string]struct {
		url                  string
		expectedMediaType    string
		expectedDownloadOnly bool
	}{
		"safe markdown": {
			url:                  "http://localhost/files/a.md",
			expectedMediaType:    "text/markdown; charset=utf-8",
			expectedDownloadOnly: false,
		},
		"large markdown": {
			url:                  "http://localhost/files/level1/large-file.md",
			expectedMediaType:    "text/markdown; charset=utf-8",
			expectedDownloadOnly: true,
		},
		"unsafe html": {
			url:                  "http://localhost/files/level1/specialFiles/javascript.html",
			expectedMediaType:    "text/html; charset=utf-8",
			expectedDownloadOnly: true,
		},
		"unsafe html without extension": {
			url:                  "http://localhost/files/level1/specialFiles/html-without-extension",
			expectedMediaType:    "text/html; charset=utf-8",
			expectedDownloadOnly: true,
		},
	}

	// Given
	server := createTestServer(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			res, err := server.Client().Get(tc.url)
			require.NoError(t, err)

			// Then
			actualDownloadOnly := res.Header.Get("Content-Disposition") != ""

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, tc.expectedMediaType, res.Header.Get("Content-Type"))
			// Prevent the browser from doing MIME-type sniffing and accept only the type sent.
			assert.Equal(t, "nosniff", res.Header.Get("X-Content-Type-Options"))
			assert.Equal(t, tc.expectedDownloadOnly, actualDownloadOnly)
		})
	}
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

func Test_should_return_not_found_for_nonexistent_resource(t *testing.T) {
	// Given
	server := createTestServer(t)

	// When
	res, err := server.Client().Get("http://localhost/files/A/B/C/DOES-NOT-EXIST.md")
	require.NoError(t, err)

	// Then
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	require.NoError(t, err)

	// Has a breadcrumb.
	actualBreadcrumb := doc.Find("nav li > a")
	assert.Greater(t, actualBreadcrumb.Length(), 1)

	// Tells the user which file was not found.
	missingFileText := doc.Find("section > p").First().Text()
	assert.Contains(t, missingFileText, "DOES-NOT-EXIST.md")

	// Lets the user return to the parent
	turnBackLink := doc.Find("a:contains('Turn back.')")
	href, _ := turnBackLink.Attr("href")
	assert.Equal(t, "files/A/B/C", href)
}

func Test_http_security_headers(t *testing.T) {
	testCases := map[string]struct {
		url string
	}{
		"folder": {url: "http://localhost/files"},
		"file":   {url: "http://localhost/files/a.md"},
		"asset":  {url: "http://localhost/static/css/ls.css"},
	}
	// Given
	server := createTestServer(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When
			res, err := server.Client().Get(tc.url)
			require.NoError(t, err)

			// Then
			// Test selected headers to verify that the middleware is installed.
			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, "DENY", res.Header.Get("X-Frame-Options"))
			assert.Equal(t, "same-site", res.Header.Get("Cross-Origin-Resource-Policy"))
		})
	}
}

func Test_should_have_html_base_path(t *testing.T) {
	testCases := map[string]struct {
		baseURL          string
		expectedBasePath string
	}{
		"default":      {baseURL: "/", expectedBasePath: "/"},
		"behind proxy": {baseURL: "/ls/", expectedBasePath: "/ls/"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			server := createTestServerWithPath(t, tc.baseURL)
			// When
			res, err := server.Client().Get("http://localhost/files")
			require.NoError(t, err)

			// Then
			doc, err := goquery.NewDocumentFromReader(res.Body)
			require.NoError(t, err)
			actualBasePath := doc.Find("head > base")
			href, _ := actualBasePath.Attr("href")
			assert.Equal(t, tc.expectedBasePath, href)
		})
	}
}

func Test_should_have_a_turn_back_link_for_empty_folders(t *testing.T) {
	// Given
	server := createTestServer(t)

	// When
	res, err := server.Client().Get("http://localhost/files/level1/level2/emptyDir")
	require.NoError(t, err)

	// Then
	assert.Equal(t, http.StatusOK, res.StatusCode)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	require.NoError(t, err)

	// Lets the user return to the parent
	turnBackLink := doc.Find("a:contains('Turn back.')")
	href, _ := turnBackLink.Attr("href")
	assert.Equal(t, "files/level1/level2", href)
}

func Test_should_provide_downloads_as_zip(t *testing.T) {
	testCases := map[string]struct {
		paths             []string
		expectedFileNames []string
	}{
		"single file": {
			paths:             []string{"a.md"},
			expectedFileNames: []string{"a.md"},
		},
		"folder": {
			paths: []string{"level1"},
			expectedFileNames: []string{
				"large-file.md",
				"level2/",
				"level2/emptyDir/",
				"specialFiles/",
				"specialFiles/<a href=\"google.com\">Link file",
				"specialFiles/folder#fragment/",
				"specialFiles/folder#fragment/.gitkeep",
				"specialFiles/folder?query=2/",
				"specialFiles/folder?query=2/.gitkeep",
				"specialFiles/html-without-extension",
				"specialFiles/javascript.html",
			},
		},
		"multiple files selected": {
			paths:             []string{"a.md", "level1/specialFiles/javascript.html"},
			expectedFileNames: []string{"a.md", "level1/specialFiles/javascript.html"},
		},
		"folder and file selected": {
			paths:             []string{"a.md", "level1/specialFiles/folder#fragment"},
			expectedFileNames: []string{"a.md", ".gitkeep"},
		},
	}
	// Given
	server := createTestServer(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When

			res, err := server.Client().PostForm("http://localhost/api/download/zip", url.Values{
				"paths": tc.paths,
			})
			require.NoError(t, err)

			// Then
			assert.Equal(t, http.StatusOK, res.StatusCode)

			zipBytes, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			bytesReader := bytes.NewReader(zipBytes)
			zipReader, err := zip.NewReader(bytesReader, int64(len(zipBytes)))
			require.NoError(t, err)

			var actualFileNames []string
			for _, file := range zipReader.File {
				actualFileNames = append(actualFileNames, file.Name)
			}

			assert.ElementsMatch(t, actualFileNames, tc.expectedFileNames)
		})
	}
}

func Test_should_fail_for_non_valid_zip_requests(t *testing.T) {
	testCases := map[string]struct {
		baseURL string
		paths   []string
	}{
		"non existing path": {
			paths: []string{"DOES-NOT-EXIST.md"},
		},
		"valid + non existing path": {
			paths: []string{"a.md", "DOES-NOT-EXIST.md"},
		},
		"path traversal": {
			paths: []string{"../lsgo_test.go"},
		},
		"absolute file linux": {
			paths: []string{"/tmp"},
		},
		"absolute file windows": {
			paths: []string{"C:\\Users\\Public"},
		},
	}
	// Given
	server := createTestServer(t)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// When

			res, err := server.Client().PostForm("http://localhost/api/download/zip", url.Values{
				"paths": tc.paths,
			})
			require.NoError(t, err)

			// Then
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		})
	}
}
