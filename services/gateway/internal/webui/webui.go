// Package webui serves the Vue 3 SPA bundled at build time via //go:embed.
// Any request that does not match a concrete file in the embedded dist is
// rewritten to /index.html so Vue Router's history mode works without a
// dedicated front-end server.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var embedded embed.FS

// Handler returns an http.Handler that serves the embedded SPA. API routes
// must be registered on the mux before mounting this handler on "/", since
// the Go mux picks the longest prefix match and "/" is the catch-all.
func Handler() http.Handler {
	root, err := fs.Sub(embedded, "dist")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		if reqPath == "" {
			files.ServeHTTP(w, r)
			return
		}
		if _, err := fs.Stat(root, reqPath); err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			files.ServeHTTP(w, r2)
			return
		}
		files.ServeHTTP(w, r)
	})
}
