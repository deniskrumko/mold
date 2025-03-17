package main

import (
	"embed"
	"log"
	"os"

	"github.com/abiosoft/mold"
)

//go:embed static
var staticDir embed.FS

func main() {
	layout, err := mold.New(staticDir, "static/layout.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := layout.Render(os.Stdout, "static/index.html", nil); err != nil {
		log.Fatal(err)
	}

	if err := layout.Render(os.Stdout, "static/hello.html", nil); err != nil {
		log.Fatal(err)
	}
}
