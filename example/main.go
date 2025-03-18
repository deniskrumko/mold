package main

import (
	"embed"
	"log"
	"os"

	"github.com/abiosoft/mold"
)

//go:embed *.html partials
var dir embed.FS
var layout = mold.New(dir)

func main() {
	if err := layout.Render(os.Stdout, "index.html", nil); err != nil {
		log.Fatal(err)
	}

	if err := layout.Render(os.Stdout, "hello.html", nil); err != nil {
		log.Fatal(err)
	}
}
