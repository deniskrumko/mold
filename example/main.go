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
	layout, err := mold.New(staticDir, "layouts/default.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := layout.Render(os.Stdout, "pages/index.html", nil); err != nil {
		log.Fatal(err)
	}

	if err := layout.Render(os.Stdout, "pages/hello.html", nil); err != nil {
		log.Fatal(err)
	}
}
