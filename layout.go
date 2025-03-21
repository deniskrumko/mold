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

func newLayout(fsys fs.FS, options *Config) (Layout, error) {
	opt, err := processConfig(fsys, options)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}
	t := &tplLayout{
		fs:   opt.fs,
		root: template.New("root"),
		set:  map[string]*template.Template{},
	}

	// traverse for all templates
	if err := t.walk(opt.exts); err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	// parse layout template
	t.layout, err = template.New("layout").Funcs(opt.funcMap).Parse(opt.layout)
	if err != nil {
		return nil, fmt.Errorf("error creating new layout: %w", err)
	}

	// process template tree for layout
	refs := processTree(nil, 0, t.layout.Tree.Root, true, true)
	for _, ref := range refs {
		if ref == bodySection || ref == headSection {
			continue
		}
		tpl := t.root.Lookup(ref)
		if tpl == nil {
			return nil, fmt.Errorf("error parsing template '%s': %w", ref, ErrNotFound)
		}
		t.layout.AddParseTree(ref, tpl.Tree)
	}

	return t, nil
}

type tplLayout struct {
	fs fs.FS

	layout *template.Template
	root   *template.Template

	set map[string]*template.Template
}

// Render implements Layout.
func (t *tplLayout) Render(w io.Writer, view string, data any) error {
	l, err := t.parseView(view)
	if err != nil {
		return err
	}

	if err := l.Execute(w, data); err != nil {
		return fmt.Errorf("error rendering layout with view '%s': %w", view, err)
	}

	return nil
}

func (t *tplLayout) parseView(name string) (*template.Template, error) {
	// reuse if previously computed
	if l, ok := t.set[name]; ok {
		return l, nil
	}

	l, err := t.layout.Clone()
	if err != nil {
		return nil, fmt.Errorf("error creating layout for view '%s': %w", name, err)
	}

	body := t.root.Lookup(name)
	if body == nil {
		return nil, ErrNotFound
	}

	// process template tree for body
	refs := processTree(nil, 0, body.Tree.Root, false, true)
	for _, ref := range refs {
		tpl := t.root.Lookup(ref)
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

	t.set[name] = l

	return l, nil
}

func (t *tplLayout) walk(exts []string) error {
	return fs.WalkDir(t.fs, ".", func(path string, d fs.DirEntry, err error) error {
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

		if nt, err := t.root.New(path).Funcs(placeholderFuncs).Parse(f); err != nil {
			return fmt.Errorf("error parsing template '%s': %w", path, err)
		} else {
			*t.root = *nt
		}

		return nil
	})
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
