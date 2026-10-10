package e2e

import (
	"archive/zip"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func appContainer(
	t *testing.T,
	baseURL string,
	options ...testcontainers.ContainerCustomizer,
) testcontainers.Container {
	t.Helper()

	absTestDataPath, err := filepath.Abs("testdata")
	require.NoError(t, err)

	command := []string{"-addr", ":8080", "-folder", "/app/testdata"}
	if baseURL != "" {
		command = append(command, "-base-url", baseURL)
	}
	containerOptions := []testcontainers.ContainerCustomizer{
		testcontainers.WithCmd(command...),
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
	}
	containerOptions = append(containerOptions, options...)

	container, err := testcontainers.Run(
		t.Context(),
		"ghcr.io/mszalbach/lsgo:0.0.0-snapshot-amd64",
		containerOptions...)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	return container
}

func nginxContainer(
	t *testing.T,
	nw *testcontainers.DockerNetwork,
	options ...testcontainers.ContainerCustomizer,
) testcontainers.Container {
	t.Helper()

	containerOptions := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("80/tcp"),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader: strings.NewReader(`server {
    listen 80;

    location = /lsgo {
        return 301 /lsgo/;
    }

    location /lsgo/ {
        proxy_pass http://app:8080/;
    }
}`),
			ContainerFilePath: "/etc/nginx/conf.d/default.conf",
			FileMode:          0o644,
		}),
		network.WithNetwork([]string{"nginx"}, nw),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("80/tcp")),
	}
	containerOptions = append(containerOptions, options...)

	container, err := testcontainers.Run(t.Context(), "nginx:alpine", containerOptions...)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	return container
}

func Test_browser_usage(t *testing.T) {
	// Setup
	container := appContainer(t, "")
	port, err := container.MappedPort(t.Context(), "8080")
	require.NoError(t, err)

	binary := launcher.New().Headless(true).NoSandbox(true)
	debugURL := binary.MustLaunch()
	browser := rod.New().ControlURL(debugURL).MustConnect().Timeout(10 * time.Second)
	t.Cleanup(browser.MustClose)

	// Tests
	t.Run("Breadcrumb Navigation", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("http://localhost:" + port.Port() + "/files/team")
		t.Cleanup(page.MustClose)

		breadcrumbs := page.MustElements("nav[aria-label='Breadcrumb'] a")
		require.Len(t, breadcrumbs, 2)
		assert.Equal(t, "Home", breadcrumbs[0].MustText())
		assert.Equal(t, "team", breadcrumbs[1].MustText())
		homeHref := breadcrumbs[0].MustAttribute("href")
		require.NotNil(t, homeHref)
		assert.Equal(t, "/files/", *homeHref)
		breadcrumbs[0].MustClick()
		page.MustWaitLoad()
		require.Len(t, page.MustElements("nav[aria-label='Breadcrumb'] a"), 1)
	})

	t.Run("Folder Content", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("http://localhost:" + port.Port())
		t.Cleanup(page.MustClose)

		readme := page.MustElement("a[href='/files/README.md']")
		assert.Equal(t, "README.md", readme.MustText())

		teamFolder := page.MustElement("a[href='/files/team']")
		assert.Equal(t, "team", teamFolder.MustText())
	})

	t.Run("Files can be sorted by name", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("http://localhost:" + port.Port())
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

		// since js is used to sort the table, we need to wait for the aria-sort attribute to be updated before checking the order of the files
		waitForSort := func(sortOrder string) func(c *assert.CollectT) {
			return func(c *assert.CollectT) {
				ariaSort := nameHeader.MustAttribute("aria-sort")
				assert.Equal(c, sortOrder, *ariaSort)
			}
		}

		assert.EventuallyWithT(t, waitForSort("ascending"), 5*time.Second, 100*time.Millisecond)
		assert.Equal(t, []string{"team", "budget.csv", "meeting-notes.md", "project-plan.md", "README.md"}, fileNames())

		nameHeader.MustClick()
		assert.EventuallyWithT(t, waitForSort("descending"), 5*time.Second, 100*time.Millisecond)
		assert.Equal(t, []string{"team", "README.md", "project-plan.md", "meeting-notes.md", "budget.csv"}, fileNames())

		nameHeader.MustClick()
		assert.EventuallyWithT(t, waitForSort("ascending"), 5*time.Second, 100*time.Millisecond)
		assert.Equal(t, []string{"team", "budget.csv", "meeting-notes.md", "project-plan.md", "README.md"}, fileNames())
	})

	t.Run("Normal file is shown in browser", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("http://localhost:" + port.Port() + "/files/README.md")
		t.Cleanup(page.MustClose)

		assert.Contains(t, page.MustElement("body").MustText(), "Shared project workspace")
	})

	t.Run("Selected files and folders can be downloaded", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("http://localhost:" + port.Port())
		t.Cleanup(page.MustClose)

		downloadDir := t.TempDir()
		waitForDownload := page.Browser().WaitDownload(downloadDir)
		page.MustElement("input[name='paths'][value='README.md']").MustClick()
		page.MustElement("input[name='paths'][value='team']").MustClick()
		page.MustElement("form button[type='submit']").MustClick()
		download := waitForDownload()
		assert.Contains(t, download.SuggestedFilename, ".zip")

		archive, err := zip.OpenReader(filepath.Join(downloadDir, download.GUID))
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, archive.Close()) })

		contents := make(map[string]string, len(archive.File))
		for _, file := range archive.File {
			reader, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.NoError(t, reader.Close())
			contents[file.Name] = string(content)
		}
		assert.Equal(t, map[string]string{
			"README.md":          "# Shared project workspace\n\nProject documents and working notes for the product team.",
			"team/":              "",
			"team/onboarding.md": "# Team onboarding\n\nStart here for the current project context and team contacts.",
			"team/roadmap.md":    "# Product roadmap\n\n- Q1: Improve the core workflow\n- Q2: Add reporting for project owners",
		}, contents)
	})
}

func Test_brwoser_behind_proxy_usage(t *testing.T) {
	// Setup
	nw, err := network.New(t.Context())
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, nw)

	appContainer(t, "/lsgo", network.WithNetwork([]string{"app"}, nw))
	proxy := nginxContainer(t, nw)
	port, err := proxy.MappedPort(t.Context(), "80")
	require.NoError(t, err)

	binary := launcher.New().Headless(true).NoSandbox(true)
	debugURL := binary.MustLaunch()
	browser := rod.New().ControlURL(debugURL).MustConnect().Timeout(10 * time.Second)
	t.Cleanup(browser.MustClose)

	// Tests
	t.Run("Breadcrumb Navigation", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("http://localhost:" + port.Port() + "/lsgo/files/team")
		t.Cleanup(page.MustClose)

		breadcrumbs := page.MustElements("nav[aria-label='Breadcrumb'] a")
		require.Len(t, breadcrumbs, 2)
		assert.Equal(t, "Home", breadcrumbs[0].MustText())
		assert.Equal(t, "team", breadcrumbs[1].MustText())
		homeHref := breadcrumbs[0].MustAttribute("href")
		require.NotNil(t, homeHref)
		assert.Equal(t, "/lsgo/files/", *homeHref)
		breadcrumbs[0].MustClick()
		page.MustWaitLoad()
		require.Len(t, page.MustElements("nav[aria-label='Breadcrumb'] a"), 1)
	})

	t.Run("Folder Content", func(t *testing.T) {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("http://localhost:" + port.Port() + "/lsgo/")
		t.Cleanup(page.MustClose)

		readme := page.MustElement("a[href='/lsgo/files/README.md']")
		assert.Equal(t, "README.md", readme.MustText())

		teamFolder := page.MustElement("a[href='/lsgo/files/team']")
		assert.Equal(t, "team", teamFolder.MustText())
	})
}
