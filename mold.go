package mold

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
)

// Layout represents a web page structure, incorporating a specific view.
// It encapsulates the body and other sections and inserts the dynamic content of a view
// into a designated area.
type Layout interface {
	// Render executes the layout template, merging it with the specified view template,
	// then writes the resulting HTML to the provided io.Writer.
	//
	// Parameters:
	//   w: The writer to which the rendered HTML will be written.
	//   view: The path of the view template whose content will be injected into the layout.
	//   data: The data to be made available to both the layout and view templates during rendering.
	//
	// Returns:
	//   An error, if any, that occurred during template execution or writing to the writer.
	Render(w io.Writer, view string, data any) error
}

// Config is the configuration for a new Layout.
type Config struct {
	// Path to the layout file.
	Layout string
	// Root subdirectory for views and partials.
	//
	// NOTE: this is not applicable to the layout path.
	Root string
	// Filename extensions for the templates. Only files with the specified extensions
	// would be parsed.
	// Default: ["html", "gohtml", "tpl", "tmpl"]
	Exts []string
	// Functions that are available for use in the templates.
	FuncMap template.FuncMap
}

// ErrNotFound is returned when a template is not found.
var ErrNotFound = errors.New("template not found")

// New creates a new Layout with fs as the underlying filesystem.
func New(fs fs.FS) (Layout, error) {
	return newLayout(fs, nil)
}

// NewWithConfig is like [New] with support for config.
func NewWithConfig(fs fs.FS, c Config) (Layout, error) {
	return newLayout(fs, &c)
}

// Must is a helper that wraps a call to a function returning ([Layout], error)
// and panics if the error is non-nil.
//
//	var t = mold.Must(mold.New(fs))
func Must(l Layout, err error) Layout {
	if err != nil {
		panic(err)
	}
	return l
}
