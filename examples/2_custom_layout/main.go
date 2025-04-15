package main

import (
	"embed"
	"net/http"

	"github.com/deniskrumko/mold"
)

//go:embed web
var dir embed.FS

var options = mold.With(
	mold.WithRoot("web"),
	mold.WithLayout("layouts/layout.html"),
)
var engine = mold.Must(mold.New(dir, options))

func main() {
	http.Handle("/", http.HandlerFunc(handleIndex))
	http.Handle("/noscript", http.HandlerFunc(handleNoScript))
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	engine.Render(w, "pages/index.html", nil)
}

func handleNoScript(w http.ResponseWriter, r *http.Request) {
	engine.Render(w, "pages/noscript.html", nil)
}
