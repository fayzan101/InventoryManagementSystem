package swagger

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed ui/*
var uiFS embed.FS

var spec []byte

// Register mounts Swagger UI and the OpenAPI spec (public, no auth).
func Register(mux *http.ServeMux, openAPISpec []byte) {
	spec = openAPISpec
	mux.HandleFunc("/swagger", redirectToUI)
	mux.HandleFunc("/swagger/", serveUI)
	mux.HandleFunc("/swagger/openapi.yaml", serveOpenAPISpec)
}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec)
}

func redirectToUI(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
}

func serveUI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/swagger/")
	if path == "" {
		path = "index.html"
	}

	content, err := fs.ReadFile(uiFS, "ui/"+path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
