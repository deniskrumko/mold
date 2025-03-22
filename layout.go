package mold

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

// defaults
var (
	//go:embed layout.html
	defaultLayout string

	// default filename extenstions for template files
	defaultExts = []string{"html", "gohtml", "tpl", "tmpl"}
)

const (
	// sections
	bodySection = "body"
	headSection = "head"
)

type (
	templateSet map[string]*template.Template
	moldEngine  templateSet
)

func newEngine(fsys fs.FS, c *Config) (Engine, error) {
	m := moldEngine{}

	opt, err := setup(fsys, c)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	// traverse to fetch all templates and populate the root template.
	root, ts, err := m.walk(opt.fs, opt.exts)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	// process layout
	layout, err := parseLayout(root, opt.layout, opt.funcMap)
	if err != nil {
		return nil, fmt.Errorf("error parsing layout: %w", err)
	}

	// process views
	for _, t := range ts {
		// ignore layout file
		if t.name == opt.layout.name {
			continue
		}
		view, err := parseView(root, layout, t.name, t.body)
		if err != nil {
			return nil, err
		}
		m[t.name] = view
	}

	return m, nil
}

// Render implements Layout.
func (m moldEngine) Render(w io.Writer, view string, data any) error {
	layout, ok := m[view]
	if !ok {
		return ErrNotFound
	}

	if err := layout.Execute(w, data); err != nil {
		return fmt.Errorf("error rendering '%s': %w", view, err)
	}

	return nil
}

func (m moldEngine) walk(fsys fs.FS, exts []string) (root templateSet, ts []templateFile, err error) {
	root = templateSet{}
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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

		f, err := readFile(fsys, path)
		if err != nil {
			return err
		}

		if t, err := template.New(path).Funcs(placeholderFuncs).Parse(f); err != nil {
			return fmt.Errorf("error parsing template '%s': %w", path, err)
		} else {
			root[path] = t
			ts = append(ts, templateFile{name: path, body: f})
		}

		return nil
	})

	return
}

type templateFile struct {
	name string
	body string
}

func setup(fsys fs.FS, c *Config) (conf struct {
	layout  templateFile
	funcMap template.FuncMap
	exts    []string
	cache   bool
	fs      fs.FS
}, err error) {
	// defaults
	conf.layout.body = defaultLayout
	conf.layout.name = "default_layout"
	conf.exts = defaultExts
	conf.fs = fsys
	conf.funcMap = placeholderFuncs

	if c == nil {
		return conf, nil
	}
	if c.root != "" {
		sub, err := fs.Sub(fsys, c.root)
		if err != nil {
			return conf, fmt.Errorf("error setting subdirectory '%s': %w", c.root, err)
		}
		conf.fs = sub
	}
	if c.layout != "" {
		f, err := readFile(conf.fs, c.layout)
		if err != nil {
			return conf, fmt.Errorf("error reading layout file '%s': %w", c.layout, err)
		}
		conf.layout.body = f
		conf.layout.name = c.layout
	}
	if len(c.exts) > 0 {
		conf.exts = c.exts
	}
	for k, f := range c.funcMap {
		conf.funcMap[k] = f
	}

	return conf, nil
}

func parseLayout(root templateSet, l templateFile, funcMap template.FuncMap) (*template.Template, error) {
	layout, err := template.New("layout").Funcs(funcMap).Parse(l.body)
	if err != nil {
		return nil, err
	}

	// process template tree for layout
	refs, err := processTree(layout, l.body, true, true)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}
	for _, ref := range refs {
		t := root[ref.name]
		if t == nil {
			if ref.typ == partialFunc {
				return nil, fmt.Errorf("error parsing template '%s': %w", ref.name, ErrNotFound)
			}
			t, _ = template.New(ref.name).Parse("")
		}
		layout.AddParseTree(ref.name, t.Tree)
	}

	return layout, nil
}

func parseView(root templateSet, layout *template.Template, name, raw string) (*template.Template, error) {
	view, err := layout.Clone()
	if err != nil {
		return nil, fmt.Errorf("error creating layout for view '%s': %w", name, err)
	}

	body := root[name]
	if body == nil {
		return nil, ErrNotFound
	}

	// process template tree for body
	refs, err := processTree(body, raw, false, true)
	if err != nil {
		return nil, fmt.Errorf("error parsing view '%s': %w", name, err)
	}
	for _, ref := range refs {
		t := root[ref.name]
		if t == nil {
			return nil, fmt.Errorf("error parsing template '%s': %w", ref.name, ErrNotFound)
		}
		view.AddParseTree(ref.name, t.Tree)
	}

	// add defined templates to the layout
	for _, t := range body.Templates() {
		tName := t.Name()
		if tName == name {
			tName = bodySection
		}
		view.AddParseTree(tName, t.Tree)
	}

	return view, nil
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

var placeholderFuncs = map[string]any{
	renderFunc.String():  func(...string) string { return "" },
	partialFunc.String(): func(string, ...any) string { return "" },
}
