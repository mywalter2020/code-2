package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*
var webFS embed.FS

func (s *Server) registerUI(mux *http.ServeMux) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/ui/", http.StripPrefix("/ui/", fileServer))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})
}
