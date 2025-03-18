package mold

import (
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

// Options is the configuration options for creation a new Layout.
type Options struct {
	// Path to the layout file.
	Layout string
	// Root directory for views and partials.
	// NOTE: this is not applicable to the layout path.
	Root string
	// If set to true, templates would be read from disk and parsed on each request.
	// Useful for quick feedback during development, otherwise should left as false.
	NoCache bool
}

// New creates a new Layout with fs as the underlying filesystem.
func New(fs fs.FS) (Layout, error) {
	return newLayout(fs, nil)
}

// NewWithOptions is like [New] with support for options.
func NewWithOptions(fs fs.FS, options Options) (Layout, error) {
	return newLayout(fs, &options)
}
