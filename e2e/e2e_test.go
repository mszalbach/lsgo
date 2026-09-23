package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func appContainer(t *testing.T) testcontainers.Container {
	t.Helper()

	absTestDataPath, err := filepath.Abs("testdata")
	require.NoError(t, err)

	container, err := testcontainers.Run(
		t.Context(), "ghcr.io/mszalbach/lsgo:0.0.0-snapshot-amd64",
		testcontainers.WithCmd("-addr", ":8080", "-folder", "/app/testdata"),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      absTestDataPath,
			ContainerFilePath: "/app/testdata",
			FileMode:          0o755,
		}),
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("8080/tcp"),
			wait.ForLog("Serving folder"),
		),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	return container
}

func Test_browser_usage(t *testing.T) {
	// Setup
	container := appContainer(t)
	port, err := container.MappedPort(t.Context(), "8080")
	require.NoError(t, err)

	baseURL := "http://localhost:" + port.Port()
	binary := launcher.New().Headless(true).NoSandbox(true)
	debugURL := binary.MustLaunch()
	browser := rod.New().ControlURL(debugURL).MustConnect().Timeout(10 * time.Second)
	t.Cleanup(browser.MustClose)

	// Tests
	t.Run("Breadcrumb Navigation", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(baseURL + "/files/folder")
		t.Cleanup(page.MustClose)

		breadcrumbs := page.MustElements("nav[aria-label='Breadcrumb'] a")
		require.Len(t, breadcrumbs, 2)
		assert.Equal(t, "Home", breadcrumbs[0].MustText())
		assert.Equal(t, "folder", breadcrumbs[1].MustText())
		homeHref := breadcrumbs[0].MustAttribute("href")
		require.NotNil(t, homeHref)
		assert.Equal(t, "files/", *homeHref)
		breadcrumbs[0].MustClick()
		page.MustWaitLoad()
		require.Len(t, page.MustElements("nav[aria-label='Breadcrumb'] a"), 1)
	})

	t.Run("Folder Content", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(baseURL)
		t.Cleanup(page.MustClose)

		helloFile := page.MustElement("a[href='files/hello.md']")
		assert.Equal(t, "hello.md", helloFile.MustText())

		folder := page.MustElement("a[href='files/folder']")
		assert.Equal(t, "folder", folder.MustText())
	})

	t.Run("Files can be sorted by name", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(baseURL)
		t.Cleanup(page.MustClose)

		nameHeader := page.MustElement("th[data-sort-key='name']")
		fileNames := func() []string {
			links := page.MustElements("tbody tr td[data-sort-column='name'] a")
			names := make([]string, 0, len(links))
			for _, link := range links {
				names = append(names, link.MustText())
			}
			return names
		}

		ariaSort := nameHeader.MustAttribute("aria-sort")
		require.NotNil(t, ariaSort)
		assert.Equal(t, "ascending", *ariaSort)
		assert.Equal(t, []string{"folder", "alpha.md", "hello.md", "zeta.md"}, fileNames())

		nameHeader.MustClick()
		ariaSort = nameHeader.MustAttribute("aria-sort")
		require.NotNil(t, ariaSort)
		assert.Equal(t, "descending", *ariaSort)
		assert.Equal(t, []string{"folder", "zeta.md", "hello.md", "alpha.md"}, fileNames())

		nameHeader.MustClick()
		ariaSort = nameHeader.MustAttribute("aria-sort")
		require.NotNil(t, ariaSort)
		assert.Equal(t, "ascending", *ariaSort)
		assert.Equal(t, []string{"folder", "alpha.md", "hello.md", "zeta.md"}, fileNames())
	})

	t.Run("Normal file is shown in browser", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(baseURL + "/files/hello.md")
		t.Cleanup(page.MustClose)

		assert.Contains(t, page.MustElement("body").MustText(), "Top file")
	})

	t.Run("File can be downloaded with the download button", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(baseURL)
		t.Cleanup(page.MustClose)

		downloadDir := t.TempDir()
		waitForDownload := page.Browser().WaitDownload(downloadDir)
		page.MustElement("a[aria-label='Download hello.md']").MustClick()
		download := waitForDownload()
		require.Equal(t, "hello.md", download.SuggestedFilename)

		downloaded, err := os.ReadFile(filepath.Join(downloadDir, download.GUID))
		require.NoError(t, err)
		assert.Equal(t, "Top file", string(downloaded))
	})

	t.Run("Unsafe html file is provided as download", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(baseURL + "/files/folder")
		t.Cleanup(page.MustClose)

		downloadDir := t.TempDir()
		waitForDownload := page.Browser().WaitDownload(downloadDir)
		page.MustElement("a[href='files/folder/javascript.html']").MustClick()
		download := waitForDownload()
		require.Equal(t, "javascript.html", download.SuggestedFilename)

		downloaded, err := os.ReadFile(filepath.Join(downloadDir, download.GUID))
		require.NoError(t, err)
		assert.Equal(
			t,
			"<!DOCTYPE html>\n<html>\n    <script>console.log(\"Hello\")</script>\n</html>",
			string(downloaded),
		)
	})

	t.Run("Not existing files produce a warning", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage(baseURL + "/files/folder/NOT-EXISTS")
		t.Cleanup(page.MustClose)

		bodyText := page.MustElement("body").MustText()
		assert.Contains(t, bodyText, "NOT-EXISTS does not exist")

		turnBackLink := page.MustElement("section a[href='files/folder']")
		assert.Equal(t, "Turn back.", turnBackLink.MustText())
	})
}
