package main

import (
	"embed"
	"net/http"

	"github.com/abiosoft/mold"
)

//go:embed web
var dir embed.FS

var layout = mold.Must(mold.New(dir,
	mold.WithRoot("web"),
	mold.WithLayout("layouts/layout.html"),
))

func main() {
	http.Handle("/", http.HandlerFunc(handleIndex))
	http.Handle("/noscript", http.HandlerFunc(handleNoScript))
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	layout.Render(w, "pages/index.html", nil)
}

func handleNoScript(w http.ResponseWriter, r *http.Request) {
	layout.Render(w, "pages/noscript.html", nil)
}
