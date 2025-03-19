package mold

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

//go:embed layout.html
var defaultLayout string

func newLayout(fsys fs.FS, options *Options) (Layout, error) {
	t := &tplLayout{
		fs:    fsys,
		views: map[string]*template.Template{},
		parts: map[string]*template.Template{},
	}
	funcs := map[string]any{"partial": t.renderPartial}

	file := defaultLayout
	if options != nil {
		t.cache = !options.NoCache
		if options.Layout != "" {
			f, err := readFile(fsys, options.Layout)
			if err != nil {
				return nil, fmt.Errorf("error creating new layout: %w", err)
			}
			file = f
		}
		if options.Root != "" {
			sub, err := fs.Sub(fsys, options.Root)
			if err != nil {
				return nil, fmt.Errorf("error setting subdirectory '%s': %w", options.Root, err)
			}
			t.fs = sub
		}
		for k, f := range options.FuncMap {
			if k == "partial" {
				continue
			}
			funcs[k] = f
		}
	}

	l, err := template.New("layout").Funcs(funcs).Parse(file)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	t.l = l

	return t, nil
}

type tplLayout struct {
	fs fs.FS
	l  *template.Template

	cache bool
	views map[string]*template.Template
	parts map[string]*template.Template
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

	var tpl *template.Template

	if t.cache {
		if p, ok := t.parts[name]; ok {
			tpl = p
		}
	} else {
		f, err := readFile(t.fs, name)
		if err != nil {
			return "", fmt.Errorf("error rendering partial '%s': %w", name, err)
		}
		tpl, err = template.New("partial").Parse(f)
		if err != nil {
			return "", fmt.Errorf("error rendering partial '%s': %w", name, err)
		}
		t.parts[name] = tpl
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
	// it is safe to do this, template.ParseFS does same thing.
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(b), nil
}
