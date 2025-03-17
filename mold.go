package mold

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

// New creates a new layout using the undelying filesystem.
func New(fs fs.FS, layoutFile string) (Layout, error) {
	l, err := newTemplate(fs, layoutFile)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	return &tplLayout{
		l:  l,
		fs: fs,
	}, nil
}

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

type tplLayout struct {
	fs fs.FS
	l  *template.Template
}

// Render implements Layout.
func (t *tplLayout) Render(w io.Writer, view string, data any) error {
	tpl, err := parseView(t.fs, view)
	if err != nil {
		return err
	}

	head, err := execTemplate(tpl.Lookup("head"), data)
	if err != nil {
		return fmt.Errorf("error executing template head for view '%s': %w", view, err)
	}

	body, err := execTemplate(tpl, data)
	if err != nil {
		return fmt.Errorf("error executing template body for view '%s': %w", view, err)
	}

	layoutData := map[string]template.HTML{
		"head": head,
		"body": body,
	}

	if err := t.l.Execute(w, layoutData); err != nil {
		return fmt.Errorf("error rendering layout with view '%s': %w", view, err)
	}

	return nil
}

func newTemplate(fs fs.FS, name string) (*template.Template, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	tpl, err := template.New("").Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("error parseing template '%s': %w", name, err)
	}
	return tpl, nil
}

func parseView(fs fs.FS, name string) (*template.Template, error) {
	tpl, err := newTemplate(fs, name)
	if err != nil {
		return nil, fmt.Errorf("error parsing view '%s': %w", name, err)
	}

	// check for head or put an empty placeholder if missing
	if tpl.Lookup("head") == nil {
		tpl.New("head").Parse("")
	}

	return tpl, nil
}

func execTemplate(t *template.Template, data any) (template.HTML, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return template.HTML(buf.String()), nil
}
