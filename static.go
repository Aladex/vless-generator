package main

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed web/static/*
var staticFiles embed.FS

//go:embed templates/*.json
var configTemplates embed.FS

//go:embed web/templates/*.html
var htmlTemplates embed.FS

// embeddedFileServer creates an HTTP handler for embedded static files
func embeddedFileServer() http.Handler {
	// Get the sub-filesystem for web/static
	staticFS, err := fs.Sub(staticFiles, "web/static")
	if err != nil {
		panic("failed to create static file system: " + err.Error())
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clean the path and remove leading slash
		cleanPath := path.Clean(strings.TrimPrefix(r.URL.Path, "/static/"))

		// Set appropriate content type based on file extension
		switch path.Ext(cleanPath) {
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		// Serve the file from embedded filesystem
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})
}
