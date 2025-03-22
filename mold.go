package mold

import (
	"errors"
	"html/template"
	"io"
	"io/fs"
)

// Engine represents a web page renderer, incorporating a specific layout.
// It provides a flexible way to generate web pages by combining views with a
// predefined layout structure.
//
// The layout defines the overall page structure,
// while views provide the dynamic content.
type Engine interface {
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

// Config is the configuration for a new [Engine].
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
// Example:
//
//	//go:embed web
//	var dir embed.FS
//	var engine, err = mold.New(dir)
//
// To use a directory on the filesystem.
//
//	var dir = os.DirFS("web")
//	var engine, err = mold.New(dir)
//
// To specify option(s) e.g. a custom layout file.
//
//	option := mold.WithLayout("layout.html")
//	engine, err := mold.New(fs, option)
func New(fs fs.FS, options ...Option) (Engine, error) {
	if len(options) == 0 {
		return newEngine(fs, nil)
	}

	var c Config
	for _, opt := range options {
		opt(&c)
	}
	return newEngine(fs, &c)
}

// Must is a helper that wraps a call to a function returning ([Engine], error)
// and panics if the error is non-nil.
//
// Example:
//
//	engine := mold.Must(mold.New(fs))
func Must(l Engine, err error) Engine {
	if err != nil {
		panic(err)
	}
	return l
}

// With is a helper for specifying multiple options.
//
// Example:
//
//	options := mold.With(
//	    mold.WithLayout("layout.html"),
//	    mold.WithRoot("web"),
//	    mold.WithExt("html"),
//	)
//	engine := mold.New(dir, options)
func With(opts ...Option) Option {
	return func(c *Config) {
		for _, opt := range opts {
			opt(c)
		}
	}
}

// WithRoot configures the base directory from which template files are loaded.
func WithRoot(subdir string) Option { return func(c *Config) { c.root = subdir } }

// WithLayout configures the path to the layout file.
func WithLayout(layout string) Option { return func(c *Config) { c.layout = layout } }

// WithExt configures the filename extensions for the templates.
// Only files with the specified extensions would be parsed.
//
//	Default: ["html", "gohtml", "tpl", "tmpl"]
func WithExt(exts ...string) Option { return func(c *Config) { c.exts = exts } }

// WithFuncMap configures the custom Go template functions.
func WithFuncMap(funcMap template.FuncMap) Option { return func(c *Config) { c.funcMap = funcMap } }
