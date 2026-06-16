package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// RegisterStatic registers the SPA fallback handler.
// fsys should be the embedded webroot FS from main.go.
// If nil (dev mode), no static serving is registered.
func RegisterStatic(r *mux.Router) {
	// No-op: static FS is wired by RegisterStaticFS from main.go after build.
}

func RegisterStaticFS(r *mux.Router, fsys fs.FS) {
	if fsys == nil {
		return
	}
	fileServer := http.FileServer(http.FS(fsys))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := fsys.Open(path)
		if err != nil {
			// SPA fallback: serve index.html
			r.URL.Path = "/"
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	})
}
