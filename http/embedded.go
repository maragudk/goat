package http

import (
	"embed"
	"net/http"
	"strings"

	goohttp "maragu.dev/goo/http"
)

func Embedded(r *goohttp.Router, public embed.FS) {
	h := http.FileServerFS(public)
	r.Mux.Get(`/embedded/*`, func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.Replace(r.URL.Path, "embedded", "public", 1)
		h.ServeHTTP(w, r)
	})
}
