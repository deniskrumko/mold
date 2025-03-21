package main

import (
	"embed"
	"net/http"

	"github.com/abiosoft/mold"
)

//go:embed web
var dir embed.FS

var layout = mold.Must(mold.NewWithConfig(dir, mold.Config{
	Layout: "web/layouts/layout.html",
	Root:   "web",
}))

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
