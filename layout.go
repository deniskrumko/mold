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

var (
	//go:embed layout.html
	defaultLayout string

	// default filename extenstions for template files
	defaultExts = []string{"html", "gohtml", "tpl", "tmpl"}
)

const (
	bodySection = "body"
	headSection = "head"

	renderFunc  = "render"
	partialFunc = "partial"
)

type tplLayout map[string]*template.Template

func newLayout(fsys fs.FS, options *Config) (Layout, error) {
	t := tplLayout{}

	opt, err := processConfig(fsys, options)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	root := template.New("root")

	// traverse for all templates
	tpls, err := t.walk(fsys, root, opt.exts)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	// parse layout template
	layout, err := template.New("layout").Funcs(opt.funcMap).Parse(opt.layout)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	// process template tree for layout
	refs, err := processTree(layout, opt.layout, true, true)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}
	for _, ref := range refs {
		if ref == bodySection || ref == headSection {
			continue
		}
		tpl := root.Lookup(ref)
		if tpl == nil {
			return nil, fmt.Errorf("error parsing template '%s': %w", ref, ErrNotFound)
		}
		layout.AddParseTree(ref, tpl.Tree)
	}

	// process views
	for _, tpl := range tpls {
		view, err := parseView(layout, root, tpl.name, tpl.body)
		if err != nil {
			return nil, err
		}
		t[tpl.name] = view
	}

	return t, nil
}

// Render implements Layout.
func (t tplLayout) Render(w io.Writer, view string, data any) error {
	l, ok := t[view]
	if !ok {
		return ErrNotFound
	}

	if err := l.Execute(w, data); err != nil {
		return fmt.Errorf("error rendering '%s': %w", view, err)
	}

	return nil
}

func parseView(layout, root *template.Template, name, raw string) (*template.Template, error) {
	l, err := layout.Clone()
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
		l.AddParseTree(ref, tpl.Tree)
	}

	// add defined templates to the layout
	for _, tpl := range body.Templates() {
		tplName := tpl.Name()
		if tplName == name {
			tplName = bodySection
		}
		l.AddParseTree(tplName, tpl.Tree)
	}

	// check for head or put an empty placeholder if missing
	if l.Lookup("head") == nil {
		l.New("head").Parse("")
	}

	return l, nil
}

func (t *tplLayout) walk(fsys fs.FS, root *template.Template, exts []string) (ts []struct {
	name string
	body string
}, err error) {
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

func processConfig(fsys fs.FS, c *Config) (conf struct {
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
