package main

import (
	"embed"
	"log"
	"os"

	"github.com/abiosoft/mold"
)

//go:embed layouts pages
var staticDir embed.FS

func main() {
	options := mold.Options{Root: "pages"}
	layout, err := mold.NewWithOptions(staticDir, "layouts/default.html", options)
	if err != nil {
		log.Fatal(err)
	}

	if err := layout.Render(os.Stdout, "index.html", nil); err != nil {
		log.Fatal(err)
	}

	if err := layout.Render(os.Stdout, "hello.html", nil); err != nil {
		log.Fatal(err)
	}
}
