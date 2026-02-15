package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist
var frontendFiles embed.FS

// spaHandler serves the embedded frontend files. For any path that does not
// match a real file it falls back to index.html so that client-side routing
// works correctly.
func spaHandler() http.Handler {
	sub, err := fs.Sub(frontendFiles, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never serve the SPA for API paths — let them 404 normally.
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
			http.NotFound(w, r)
			return
		}
		// Try to open the requested file.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if f, err := sub.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// File not found — serve index.html for SPA routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
