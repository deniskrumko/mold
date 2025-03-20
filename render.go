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

const (
	bodySection = "body"
	headSection = "head"
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

	// include referenced templates for the layout
	refs := fetchRefs(nil, 0, t.layout.Tree.Root, true, true)
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

	// add referenced templates for the view
	refs := fetchRefs(nil, 0, body.Tree.Root, false, true)
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

// fetchRefs fetchs all templates referenced by the root note.
func fetchRefs(parent *parse.ListNode, index int, node parse.Node, render, partial bool) []string {
	var ts []string
	if a, ok := node.(*parse.ActionNode); ok {
		if len(a.Pipe.Cmds) > 0 {
			funcName, tname := getFunctionName(a.Pipe.Cmds[0])
			if funcName == "render" || funcName == "partial" {
				processActionNode(parent, index, &node, render, partial)
			}
			if tname != "" {
				ts = append(ts, tname)
			}
		}
	}
	if l, ok := node.(*parse.ListNode); ok {
		for i, n := range l.Nodes {
			ts = append(ts, fetchRefs(l, i, n, render, partial)...)
		}
	}
	if i, ok := node.(*parse.IfNode); ok {
		ts = append(ts, fetchRefs(parent, index, i.List, render, partial)...)
		ts = append(ts, fetchRefs(parent, index, i.ElseList, render, partial)...)
	}
	if r, ok := node.(*parse.RangeNode); ok {
		ts = append(ts, fetchRefs(parent, index, r.List, render, partial)...)
		ts = append(ts, fetchRefs(parent, index, r.ElseList, render, partial)...)
	}

	return ts
}

func processActionNode(parent *parse.ListNode, index int, node *parse.Node, render, partial bool) {
	if parent == nil {
		// this must never happen
		panic("parent node is nil")
	}

	actionNode := (*node).(*parse.ActionNode)
	cmd := actionNode.Pipe.Cmds[0]
	funcName, name := getFunctionName(cmd)

	switch funcName {
	case "partial":
		if !partial {
			return
		}
	case "render":
		if !render {
			return
		}
		if name == "" {
			name = "body"
		}

	default:
		return
	}

	cmd.Args = []parse.Node{&parse.DotNode{}}
	actionNode.Pipe.Cmds = []*parse.CommandNode{cmd}

	tn := &parse.TemplateNode{
		NodeType: parse.NodeTemplate,
		Pos:      actionNode.Pos,
		Line:     actionNode.Line,
		Name:     name,
		Pipe:     actionNode.Pipe,
	}

	parent.Nodes[index] = tn
}

func getFunctionName(cmd *parse.CommandNode) (fn string, file string) {
	if len(cmd.Args) > 0 {
		if i, ok := cmd.Args[0].(*parse.IdentifierNode); ok {
			fn = i.Ident
		}
	}
	if len(cmd.Args) > 1 {
		if s, ok := cmd.Args[1].(*parse.StringNode); ok {
			file = s.Text
		}
	}
	return
}

var placeholderFuncs = map[string]any{
	"render":  func(...string) string { return "" },
	"partial": func(string, ...any) string { return "" },
}
