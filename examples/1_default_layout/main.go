package main

import (
	"embed"
	"log"
	"net/http"
	"strings"

	"github.com/deniskrumko/mold"
)

//go:embed *.html partials
var dir embed.FS
var engine = mold.Must(mold.New(dir))

func main() {
	handler := http.HandlerFunc(handler)

	http.Handle("/", handler)
	http.Handle("/hello", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name string
	}
	if strings.HasPrefix(r.URL.Path, "/hello") {
		data.Name = r.FormValue("name")
		if data.Name == "" {
			data.Name = "friend"
		}
	}

	if err := engine.Render(w, "index.html", data); err != nil {
		log.Println("error during render:", err)
	}
}
