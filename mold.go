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

// Config is the configuration for a new [Layout].
type Config struct {
	root    string
	layout  string
	exts    []string
	funcMap template.FuncMap
}

// Option is a configuration option.
type Option func(*Config)

// ErrNotFound is returned when a template is not found.
var ErrNotFound = errors.New("template not found")

// New creates a new Layout with fs as the underlying filesystem.
//
//	//go:embed web
//	var dir embed.FS
//	layout, err := mold.New(dir)
//
// To use a directory on the filesystem.
//
//	var dir os.DirFS("web")
//	layout, err := mold.New(dir)
//
// To specify options. e.g. custom layout
//
//	layout, err := mold.New(fs, mold.WithLayout("layout.html"))
func New(fs fs.FS, options ...Option) (Layout, error) {
	if len(options) == 0 {
		return newLayout(fs, nil)
	}

	var c Config
	for _, opt := range options {
		opt(&c)
	}
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

// WithRoot configures the root subdirectory.
func WithRoot(root string) Option { return func(c *Config) { c.root = root } }

// WithLayout configures the path to the layout file.
func WithLayout(layout string) Option { return func(c *Config) { c.layout = layout } }

// WithExts configures the filename extensions for the templates.
// Only files with the specified extensions would be parsed.
//
//	Default: ["html", "gohtml", "tpl", "tmpl"]
func WithExts(exts []string) Option { return func(c *Config) { c.exts = exts } }

// WithFuncMap sets custom Go template functions available for use in templates.
func WithFuncMap(funcMap template.FuncMap) Option { return func(c *Config) { c.funcMap = funcMap } }
