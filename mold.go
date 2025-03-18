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

// New creates a new Layout using fs as the underlying filesystem.
func New(fs fs.FS, layoutFile string) (Layout, error) {
	f, err := readFile(fs, layoutFile)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	t := &tplLayout{fs: fs}
	funcs := map[string]any{"partial": t.renderPartial}

	layout, err := template.New("layout").Funcs(funcs).Parse(f)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	t.l = layout

	return t, nil
}

type tplLayout struct {
	fs fs.FS
	l  *template.Template
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
