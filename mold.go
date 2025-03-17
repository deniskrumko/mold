package mold

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

// Layout represents a web page structure, incorporating a specific view.
// It encapsulates the header and body and inserts the dynamic content of a view
// into a designated area.
type Layout interface {
	// Render executes the layout template, merging it with the specified view template and data,
	// then writes the resulting HTML to the provided io.Writer.
	//
	// Parameters:
	//   w: The io.Writer to which the rendered HTML will be written.
	//   view: The path of the view template whose content will be injected into the layout.
	//   data: The data to be made available to both the layout and view templates during rendering.
	//
	// Returns:
	//   An error, if any, that occurred during template execution or writing to the io.Writer.
	Render(w io.Writer, view string, data any) error
}

// New creates a new layout using fs as the underlying filesystem.
func New(fs fs.FS, layoutFile string) (Layout, error) {
	f, err := readFile(fs, layoutFile)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	l, err := template.New("layout").Parse(f)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	return &tplLayout{
		l:  l,
		fs: fs,
	}, nil
}

type tplLayout struct {
	fs fs.FS
	l  *template.Template
}

// Render implements Layout.
func (t *tplLayout) Render(w io.Writer, view string, data any) error {
	l, err := t.parseView(t.fs, view)
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

func (t *tplLayout) parseView(fs fs.FS, name string) (*template.Template, error) {
	l, err := t.l.Clone()
	if err != nil {
		return nil, fmt.Errorf("error cloning layout for view '%s': %w", name, err)
	}

	f, err := readFile(fs, name)
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

	return l, nil
}
