package mold

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template/parse"
)

var (
	//go:embed layout.html
	defaultLayout string

	// default filename extenstions for template files
	defaultExts = []string{"html", "gohtml", "tpl", "tmpl"}
)

func newLayout(fsys fs.FS, options *Options) (Layout, error) {
	opt, err := parseOptions(fsys, options)
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

	// include referenced templates for the layout
	refs := fetchRefs(t.layout.Tree.Root)
	for _, ref := range refs {
		if ref == "body" || ref == "head" {
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

	layoutData := map[string]any{
		"data": data,
	}

	if err := l.Execute(w, layoutData); err != nil {
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

	// add referenced templates for the view
	refs := fetchRefs(body.Tree.Root)
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
			tplName = "body"
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

		if nt, err := t.root.New(path).Parse(f); err != nil {
			return fmt.Errorf("error parsing template '%s': %w", path, err)
		} else {
			*t.root = *nt
		}

		return nil
	})
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

// fetchRefs fetchs all templates referenced by the root note.
func fetchRefs(node parse.Node) []string {
	var ts []string
	if t, ok := node.(*parse.TemplateNode); ok {
		ts = append(ts, t.Name)
	}
	if l, ok := node.(*parse.ListNode); ok {
		for _, n := range l.Nodes {
			ts = append(ts, fetchRefs(n)...)
		}
	}
	if i, ok := node.(*parse.IfNode); ok {
		ts = append(ts, fetchRefs(i.List)...)
		ts = append(ts, fetchRefs(i.ElseList)...)
	}
	if r, ok := node.(*parse.RangeNode); ok {
		ts = append(ts, fetchRefs(r.List)...)
		ts = append(ts, fetchRefs(r.ElseList)...)
	}

	return ts
}
