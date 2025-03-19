package mold

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
)

// Layout represents a web page structure, incorporating a specific view.
// It encapsulates the header and body and inserts the dynamic content of a view
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

// Options is the configuration options for a new Layout.
type Options struct {
	// Path to the layout file.
	Layout string
	// Root subdirectory for views and partials.
	// NOTE: this is not applicable to the layout path.
	Root string
	// Filename extensions for the templates. Only files with the specified extensions
	// would be accessible.
	// Default: ["html", "gohtml", "tpl", "tmpl"]
	Exts []string
	// FuncMap is the [template.FuncMap] that is available for use in the templates.
	FuncMap template.FuncMap
}

// ErrNotFound is returned when a template is not found.
var ErrNotFound = errors.New("template not found")

// New creates a new Layout with fs as the underlying filesystem.
func New(fs fs.FS) Layout {
	l, err := newLayout(fs, nil)
	if err != nil {
		panic(err) // this should never happen
	}
	return l
}

// NewWithOptions is like [New] with support for options.
func NewWithOptions(fs fs.FS, options Options) (Layout, error) {
	return newLayout(fs, &options)
}
