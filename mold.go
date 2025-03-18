package mold

import (
	"bytes"
	"fmt"
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

// Options is the configuration options for creation a new Layout.
type Options struct {
	// Root directory for views and partials.
	// NOTE: this is not applicable to the layout path.
	Root string
	// If set to true, templates would be read from disk and parsed on each request.
	// Useful for quick feedback during development, otherwise should left as false.
	NoCache bool
}

// New creates a new Layout with the specified layout template
// and fs as the underlying filesystem.
func New(fs fs.FS, layout string) (Layout, error) {
	return newLayout(fs, layout, nil)
}

// NewWithOptions is like [New] with support for options.
func NewWithOptions(fs fs.FS, layout string, options Options) (Layout, error) {
	return newLayout(fs, layout, &options)
}

func newLayout(fsys fs.FS, layout string, options *Options) (Layout, error) {
	f, err := readFile(fsys, layout)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	t := &tplLayout{fs: fsys, views: map[string]*template.Template{}}

	if options != nil {
		t.cache = !options.NoCache
		if options.Root != "" {
			sub, err := fs.Sub(fsys, options.Root)
			if err != nil {
				return nil, fmt.Errorf("error setting subdirectory '%s': %w", options.Root, err)
			}
			t.fs = sub
		}
	}

	funcs := map[string]any{"partial": t.renderPartial}

	t.l, err = template.New("layout").Funcs(funcs).Parse(f)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	return t, nil
}

type tplLayout struct {
	fs fs.FS
	l  *template.Template

	cache bool
	views map[string]*template.Template
}

// Render implements Layout.
func (t *tplLayout) Render(w io.Writer, view string, data any) error {
	l, err := t.parseView(view)
	if err != nil {
		return err
	}

	layoutData := map[string]any{
		"data": data,
	}

	if err := l.Execute(w, layoutData); err != nil {
		return fmt.Errorf("error rendering layout with view '%s': %w", view, err)
	}

	return nil
}

func (t *tplLayout) parseView(name string) (*template.Template, error) {
	if t.cache {
		if l, ok := t.views[name]; ok {
			return l, nil
		}
	}

	l, err := t.l.Clone()
	if err != nil {
		return nil, fmt.Errorf("error cloning layout for view '%s': %w", name, err)
	}

	f, err := readFile(t.fs, name)
	if err != nil {
		return nil, fmt.Errorf("error reading view '%s': %w", name, err)
	}

	if _, err := l.New("body").Parse(f); err != nil {
		return nil, fmt.Errorf("error parsing view '%s': %w", name, err)
	}

	// check for head or put an empty placeholder if missing
	if l.Lookup("head") == nil {
		l.New("head").Parse("")
	}

	t.views[name] = l

	return l, nil
}

func (t *tplLayout) renderPartial(name string, params ...any) (template.HTML, error) {
	var data any
	if len(params) > 0 {
		data = params[0]
	}

	f, err := readFile(t.fs, name)
	if err != nil {
		return "", fmt.Errorf("error rendering partial '%s': %w", name, err)
	}
	tpl, err := template.New("partial").Parse(f)
	if err != nil {
		return "", fmt.Errorf("error rendering partial '%s': %w", name, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error rendering partial '%s': %w", name, err)
	}

	return template.HTML(buf.String()), nil
}

func readFile(fs fs.FS, name string) (string, error) {
	f, err := fs.Open(name)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(b), nil
}
