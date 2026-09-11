// Package web contains everyhting to render the web ui to list a directory folder
package web

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mszalbach/lsgo/internal/assets"
	"github.com/mszalbach/lsgo/internal/explorer"
)

// Server provides everything needed to serve the LSGo webpage
type Server struct {
	root         explorer.Root
	htmlRenderer *htmlRenderer
}

// NewServer creates a Server
func NewServer(root explorer.Root) (Server, error) {
	renderer, err := newHTMLRenderer()
	if err != nil {
		return Server{}, err
	}

	return Server{
		root:         root,
		htmlRenderer: renderer,
	}, nil
}

// Router constructs the handlers and bind them to the correct path to serve LSGo
func (s Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", rootHandler)
	mux.HandleFunc("GET /files/{file...}", s.lsHandler)
	mux.Handle("GET /static/", http.FileServerFS(assets.Static))
	mux.HandleFunc("GET /favicon.ico", faviconHandler)

	return mux
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/files", http.StatusMovedPermanently)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, assets.Static, "static/icons/folder-eye.svg")
}

func (s Server) lsHandler(w http.ResponseWriter, r *http.Request) {
	upath := "./" + r.PathValue("file")
	upath = filepath.Clean(upath)

	file, err := s.root.File(upath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if file.IsDir {
		s.serveDirectory(w, r, file)
		return
	}

	serveFile(w, r, file)
}

func serveFile(w http.ResponseWriter, r *http.Request, file *explorer.File) {
	osFile, err := file.AsOsFile()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer osFile.Close()

	if r.URL.Query().Get("download") != "" {
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": file.Name,
		}))
	}
	http.ServeContent(w, r, file.Name, file.ModTime, osFile)
}

func (s Server) serveDirectory(w http.ResponseWriter, r *http.Request, dir *explorer.File) {
	// folder should end with a trailing slash to simplify links and work behind a proxy
	if !strings.HasSuffix(r.URL.Path, "/") {
		target := r.URL.Path + "/"
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		//nolint:gosec // Seems to be safe against "Open redirect" but needs more testing
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		return
	}

	err := s.htmlRenderer.render(w, http.StatusOK, dir, "directory.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
