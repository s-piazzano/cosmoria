package webappui

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:dist
var Dist embed.FS

func Handler() http.Handler {
	sub, err := fs.Sub(Dist, "dist")
	if err != nil {
		log.Panicf("webappui: fs.Sub: %v", err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if len(p) > 0 && p[0] == '/' {
			p = p[1:]
		}
		if p != "" {
			if _, err := fs.Stat(sub, p); err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
