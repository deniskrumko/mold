package mold

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

var (
	//go:embed layout.html
	defaultLayout string

	defaultExts = []string{"html", "gohtml", "tpl", "tmpl"}
)

func newLayout(fsys fs.FS, options *Options) (Layout, error) {
	opt, err := parseOptions(fsys, options)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}
	t := &tplLayout{
		fs:    opt.fs,
		views: map[string]*template.Template{},
		parts: map[string]*template.Template{},
	}
	// add partial renderer
	opt.funcMap["partial"] = t.renderPartial

	// parse layout template
	t.l, err = template.New("layout").Funcs(opt.funcMap).Parse(opt.layout)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	// traverse for all other templates
	if err := t.walk(opt.exts); err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	return t, nil
}

type tplLayout struct {
	fs fs.FS
	l  *template.Template

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
	if l, ok := t.views[name]; ok {
		return l, nil
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

	if p, ok := t.parts[name]; ok {
		tpl = p
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

func (t *tplLayout) walk(exts []string) error {
	err := fs.WalkDir(t.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip hidden files and directories.
		if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		if !validExt(exts, ext) {
			return nil
		}

		f, err := readFile(t.fs, path)
		if err != nil {
			return err
		}

		if _, err := t.l.New(path).Parse(f); err != nil {
			return fmt.Errorf("error parsing template '%s': %w", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error locating templates: %w", err)
	}

	return nil
}

func parseOptions(fsys fs.FS, options *Options) (opt struct {
	layout  string
	funcMap template.FuncMap
	exts    []string
	cache   bool
	fs      fs.FS
}, err error) {
	// defaults
	opt.layout = defaultLayout
	opt.exts = defaultExts
	opt.funcMap = map[string]any{}
	opt.fs = fsys

	if options == nil {
		return opt, nil
	}

	if options.Layout != "" {
		f, err := readFile(fsys, options.Layout)
		if err != nil {
			return opt, fmt.Errorf("error reading layout file '%s': %w", options.Layout, err)
		}
		opt.layout = f
	}
	if options.Root != "" {
		sub, err := fs.Sub(fsys, options.Root)
		if err != nil {
			return opt, fmt.Errorf("error setting subdirectory '%s': %w", options.Root, err)
		}
		opt.fs = sub
	}
	if len(options.Exts) > 0 {
		opt.exts = options.Exts
	}
	for k, f := range options.FuncMap {
		opt.funcMap[k] = f
	}

	return opt, nil
}

func readFile(fsys fs.FS, name string) (string, error) {
	f, err := fs.ReadFile(fsys, name)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return string(f), nil
}

func validExt(exts []string, ext string) bool {
	if ext == "" {
		return false
	}

	for _, e := range exts {
		if strings.TrimPrefix(e, ".") == strings.TrimPrefix(ext, ".") {
			return true
		}
	}

	return false
}
