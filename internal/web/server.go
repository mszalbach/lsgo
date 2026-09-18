// Package web contains everything needed to render the web UI for listing folders.
package web

import (
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/mszalbach/lsgo/internal/assets"
	"github.com/mszalbach/lsgo/internal/filesystem"
)

// Router provides everything needed to serve the LSGo webpage.
type Router struct {
	root              filesystem.Root
	htmlRenderer      *htmlRenderer
	maxInlineFileSize int64
}

// NewRouter creates a Router.
func NewRouter(root filesystem.Root, maxInlineFileSize int64) (Router, error) {
	// TODO renderer also injecting
	renderer, err := newHTMLRenderer(assets.Templates, "html/base.tmpl")
	if err != nil {
		return Router{}, err
	}

	return Router{
		root:              root,
		htmlRenderer:      renderer,
		maxInlineFileSize: maxInlineFileSize,
	}, nil
}

// Routes constructs the handlers and binds them to the correct paths to serve LSGo.
func (s Router) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", rootHandler)
	mux.HandleFunc("GET /files/{file...}", s.lsHandler)
	mux.Handle("GET /static/", http.FileServerFS(assets.Static))
	mux.HandleFunc("GET /favicon.ico", faviconHandler)

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

		if errors.Is(err, fs.ErrNotExist) {
			err = s.htmlRenderer.render(w, http.StatusNotFound, data{Content: upath}, "base", "html/pages/404.tmpl")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if file.IsDir {
		s.serveFolder(w, file)
		return
	}

	s.serveFile(w, r, file)
}

func (s Router) serveFile(w http.ResponseWriter, r *http.Request, file *filesystem.File) {
	osFile, err := file.AsOsFile()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer osFile.Close()

	mediaType, err := detectMediaType(osFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", mediaType)

	isSafeMediaType, err := isSafeInlineMediaType(mediaType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	isDownload := r.URL.Query().Get("download") != ""
	isFileTooLarge := file.Size > s.maxInlineFileSize
	isUnsecureMediaType := !isSafeMediaType

	if isDownload || isFileTooLarge || isUnsecureMediaType {
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": file.Name,
		}))
	}

	http.ServeContent(w, r, file.Name, file.ModTime, osFile)
}

func (s Router) serveFolder(w http.ResponseWriter, dir *filesystem.File) {
	breadcrumb := createBreadcrumb(dir.RelPath)
	folderData, err := folderDataFrom(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
