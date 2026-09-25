// Package web contains everything needed to render the web UI for listing folders.
package web

import (
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mszalbach/lsgo/internal/assets"
	"github.com/mszalbach/lsgo/internal/filesystem"
)

// Router provides everything needed to serve the LSGo webpage.
type Router struct {
	root              filesystem.Root
	htmlRenderer      *HTMLRenderer
	maxInlineFileSize int64
}

// NewRouter creates a Router.
func NewRouter(root filesystem.Root, renderer *HTMLRenderer, maxInlineFileSize int64) (Router, error) {
	return Router{
		root:              root,
		htmlRenderer:      renderer,
		maxInlineFileSize: maxInlineFileSize,
	}, nil
}

// Routes constructs the handlers and binds them to the correct paths to serve LSGo.
func (s Router) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /favicon.ico", faviconHandler)
	mux.HandleFunc("GET /", rootHandler)
	mux.Handle("GET /static/", http.FileServerFS(assets.Static))
	mux.HandleFunc("GET /files/{file...}", s.lsHandler)
	mux.HandleFunc("POST /api/download/zip", s.downloadZipHandler)

	return owaspMiddleware(mux)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/files", http.StatusMovedPermanently)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, assets.Static, "static/icons/folder-eye.svg")
}

func (s Router) lsHandler(w http.ResponseWriter, r *http.Request) {
	upath := "./" + r.PathValue("file")
	upath = filepath.Clean(upath)

	file, err := s.root.File(upath)
	if err != nil {
		s.handleOpenFileError(w, &filesystem.File{RelPath: upath}, err)
		return
	}

	if file.IsDir {
		s.serveFolder(w, file)
		return
	}

	s.serveFile(w, r, file)
}

func (s Router) downloadZipHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	selectedPaths := r.PostForm["paths"]
	if len(selectedPaths) == 0 {
		http.Error(w, "No files selected", http.StatusBadRequest)
		return
	}

	files := make([]*filesystem.File, 0, len(selectedPaths))
	for _, selectedPath := range selectedPaths {
		file, err := s.root.File(selectedPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusBadRequest)
			return
		}
		files = append(files, file)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": "lsgo-" + time.Now().Format("2006-01-02_150405") + ".zip",
	}))

	err = filesystem.WriteZipArchive(w, files...)
	if err != nil {
		slog.Error("Could not write zip archive for files %s: %w", slog.Any("files", files), slog.Any("error", err))
		return
	}
}

func (s Router) handleOpenFileError(w http.ResponseWriter, file *filesystem.File, err error) {
	breadcrumb := createBreadcrumb(file.RelPath)
	var renderError error
	switch {
	case errors.Is(err, os.ErrNotExist):
		renderError = s.htmlRenderer.render(
			w,
			http.StatusNotFound,
			data{Breadcrumb: breadcrumb, Content: file.RelPath},
			"base",
			"html/pages/notfound.tmpl",
		)
	case errors.Is(err, os.ErrPermission):
		renderError = s.htmlRenderer.render(
			w,
			http.StatusForbidden,
			data{Breadcrumb: breadcrumb, Content: file.RelPath},
			"base",
			"html/pages/denied.tmpl",
		)
	default:
		s.httpError(w, file, err)
		return
	}

	if renderError != nil {
		slog.Error("Could not render error template", slog.String("file", file.RelPath), slog.Any("error", renderError))
		// zip writes directly to w, so there is nothing which can be done when an error happens. See ADR-20260925-1.
		return
	}
}

func (s Router) serveFolder(w http.ResponseWriter, dir *filesystem.File) {
	breadcrumb := createBreadcrumb(dir.RelPath)
	folderData, err := folderDataFrom(dir)
	if err != nil {
		s.httpError(w, dir, err)
		return
	}

	err = s.htmlRenderer.render(
		w,
		http.StatusOK,
		data{Breadcrumb: breadcrumb, Content: folderData},
		"base",
		"html/pages/folder.tmpl",
	)
	if err != nil {
		s.httpError(w, dir, err)
		return
	}
}

func (s Router) serveFile(w http.ResponseWriter, r *http.Request, file *filesystem.File) {
	osFile, err := file.AsOsFile()
	if err != nil {
		s.httpError(w, file, err)
		return
	}
	defer osFile.Close()

	mediaType, err := detectMediaType(osFile)
	if err != nil {
		s.httpError(w, file, err)
		return
	}
	w.Header().Set("Content-Type", mediaType)

	isSafeMediaType, err := isSafeInlineMediaType(mediaType)
	if err != nil {
		s.httpError(w, file, err)
		return
	}

	isFileTooLarge := file.Size > s.maxInlineFileSize
	isUnsecureMediaType := !isSafeMediaType

	if isFileTooLarge || isUnsecureMediaType {
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": file.Name,
		}))
	}

	http.ServeContent(w, r, file.Name, file.ModTime, osFile)
}

func (s Router) httpError(w http.ResponseWriter, file *filesystem.File, err error) {
	breadcrumb := createBreadcrumb(file.RelPath)
	renderError := s.htmlRenderer.render(
		w,
		http.StatusInternalServerError,
		data{Breadcrumb: breadcrumb, Content: fmt.Sprintf("Failed %s %s", file.RelPath, err)},
		"base",
		"html/pages/error.tmpl",
	)

	if renderError != nil {
		// give up and use standard error handling
		slog.Error("Could not render error template", slog.String("file", file.RelPath), slog.Any("error", renderError))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// owaspMiddleware sets the recommended header for security
// see https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
func owaspMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().
			Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; upgrade-insecure-requests;")
		next.ServeHTTP(w, r)
	})
}
