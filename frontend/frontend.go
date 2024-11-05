package frontend

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"encore.dev"
)

var (
	//go:embed dist
	dist embed.FS

	assets, _ = fs.Sub(dist, "dist")
	handler   = http.StripPrefix("/frontend/", http.FileServer(http.FS(assets)))
)

// Serves the frontend for development
//encore:api public raw path=/frontend/*path
func Serve(w http.ResponseWriter, req *http.Request) {
	if path := encore.CurrentRequest().PathParams.Get("path"); path == "env.js" {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprintf(w, "API_BASE_URL = %q;\n", encore.Meta().APIBaseURL.String())
		return
	}

	handler.ServeHTTP(w, req)
}
