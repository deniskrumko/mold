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

	// functions
	renderFunc  = "render"
	partialFunc = "partial"
)

type moldLayout map[string]*template.Template

func newLayout(fsys fs.FS, options *Config) (Layout, error) {
	m := moldLayout{}

	opt, err := setup(fsys, options)
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
		view, err := parseView(root, layout, t.name, t.body)
		if err != nil {
			return nil, err
		}
		m[t.name] = view
	}

	return m, nil
}

// Render implements Layout.
func (m moldLayout) Render(w io.Writer, view string, data any) error {
	l, ok := m[view]
	if !ok {
		return ErrNotFound
	}

	if err := l.Execute(w, data); err != nil {
		return fmt.Errorf("error rendering '%s': %w", view, err)
	}

	return nil
}

func (m moldLayout) walk(fsys fs.FS, exts []string) (root *template.Template, ts []struct {
	name string
	body string
}, err error) {
	root = template.New("root")
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

		if nt, err := root.New(path).Funcs(placeholderFuncs).Parse(f); err != nil {
			return fmt.Errorf("error parsing template '%s': %w", path, err)
		} else {
			*root = *nt
			ts = append(ts, struct {
				name string
				body string
			}{name: path, body: f})
		}

		return nil
	})

	return
}

func setup(fsys fs.FS, c *Config) (conf struct {
	layout  string
	funcMap template.FuncMap
	exts    []string
	cache   bool
	fs      fs.FS
}, err error) {
	// defaults
	conf.layout = defaultLayout
	conf.exts = defaultExts
	conf.fs = fsys
	conf.funcMap = placeholderFuncs

	if c == nil {
		return conf, nil
	}

	if c.Layout != "" {
		f, err := readFile(fsys, c.Layout)
		if err != nil {
			return conf, fmt.Errorf("error reading layout file '%s': %w", c.Layout, err)
		}
		conf.layout = f
	}
	if c.Root != "" {
		sub, err := fs.Sub(fsys, c.Root)
		if err != nil {
			return conf, fmt.Errorf("error setting subdirectory '%s': %w", c.Root, err)
		}
		conf.fs = sub
	}
	if len(c.Exts) > 0 {
		conf.exts = c.Exts
	}
	for k, f := range c.FuncMap {
		conf.funcMap[k] = f
	}

	return conf, nil
}

func parseLayout(root *template.Template, raw string, funcMap template.FuncMap) (*template.Template, error) {
	layout, err := template.New("layout").Funcs(funcMap).Parse(raw)
	if err != nil {
		return nil, err
	}

	// process template tree for layout
	refs, err := processTree(layout, raw, true, true)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}
	for _, ref := range refs {
		if ref == bodySection || ref == headSection {
			continue
		}
		t := root.Lookup(ref)
		if t == nil {
			return nil, fmt.Errorf("error parsing template '%s': %w", ref, ErrNotFound)
		}
		layout.AddParseTree(ref, t.Tree)
	}

	return layout, nil
}

func parseView(root, layout *template.Template, name, raw string) (*template.Template, error) {
	view, err := layout.Clone()
	if err != nil {
		return nil, fmt.Errorf("error creating layout for view '%s': %w", name, err)
	}

	body := root.Lookup(name)
	if body == nil {
		return nil, ErrNotFound
	}

	// process template tree for body
	refs, err := processTree(body, raw, false, true)
	if err != nil {
		return nil, fmt.Errorf("error parsing view '%s': %w", name, err)
	}
	for _, ref := range refs {
		tpl := root.Lookup(ref)
		if tpl == nil {
			return nil, fmt.Errorf("error parsing template '%s': %w", ref, ErrNotFound)
		}
		view.AddParseTree(ref, tpl.Tree)
	}

	// add defined templates to the layout
	for _, tpl := range body.Templates() {
		tplName := tpl.Name()
		if tplName == name {
			tplName = bodySection
		}
		view.AddParseTree(tplName, tpl.Tree)
	}

	// check for head or put an empty placeholder if missing
	if view.Lookup("head") == nil {
		view.New("head").Parse("")
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
	renderFunc:  func(...string) string { return "" },
	partialFunc: func(string, ...any) string { return "" },
}
