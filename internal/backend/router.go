package backend

import (
	"mime"
	"net/http"
	"os"

	"github.com/mszalbach/lsgo/internal/assets"
)

func Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", http.NotFound)
	mux.HandleFunc("GET /files/{file...}", http.HandlerFunc(serveLs))
	mux.Handle("GET /static/", http.FileServerFS(assets.Static))

	return mux
}

func serveLs(w http.ResponseWriter, r *http.Request) {
	upath := "./" + r.PathValue("file")

	root, err := NewRootFS("./logs")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := root.File(upath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if file.IsDir {
		serveDirectory(w, file)
		return
	}

	serveFile(w, r, file)

}

func serveFile(w http.ResponseWriter, r *http.Request, file file) {
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

func serveDirectory(w http.ResponseWriter, dir file) {

	htmlRenderer, err := newHTMLRenderer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = htmlRenderer.render(w, http.StatusOK, dir, "directory.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
