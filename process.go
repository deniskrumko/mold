package mold

import (
	"fmt"
	"text/template/parse"
)

// processTree traverses the node tree and swaps render and partial declarations with equivalent template calls.
// It returns all referenced templates encountered during the traversal.
func processTree(t *templateFile) ([]nestedFile, error) {
	ts, err := processNode(t, nil, 0, t.Tree.Root)
	if err != nil {
		if err, ok := err.(posErr); ok {
			line, col := pos(t.body, err.pos)
			return ts, fmt.Errorf("%s:%d:%d: %s: %w", t.Name(), line, col, t.typ, err)
		}
	}

	return ts, nil
}

func processNode(root *templateFile, parent *parse.ListNode, index int, node parse.Node) (ts []nestedFile, err error) {
	// appendResult appends the specified templates to the list of template names when there are no errors
	appendResult := func(t []nestedFile, err1 error) {
		if err1 != nil {
			err = err1
		}
		if err == nil {
			ts = append(ts, t...)
		}
	}

	if a, ok := node.(*parse.ActionNode); ok {
		if len(a.Pipe.Cmds) > 0 {
			funcName, tname, _ := getActionArgs(a.Pipe.Cmds[0])
			if err := processActionNode(root, parent, index, node, funcName); err != nil {
				return ts, err
			}
			if funcName == partialFunc.String() && tname != "" {
				ts = append(ts, nestedFile{name: tname, typ: partialFunc})
			} else if funcName == renderFunc.String() && tname != "" {
				ts = append(ts, nestedFile{name: tname, typ: renderFunc})
			}
		}
	}

	if w, ok := node.(*parse.WithNode); ok && w != nil {
		appendResult(processNode(root, parent, index, w.List))
		appendResult(processNode(root, parent, index, w.ElseList))
	}
	if l, ok := node.(*parse.ListNode); ok && l != nil {
		for i, n := range l.Nodes {
			appendResult(processNode(root, l, i, n))
		}
	}
	if i, ok := node.(*parse.IfNode); ok && i != nil {
		appendResult(processNode(root, parent, index, i.List))
		appendResult(processNode(root, parent, index, i.ElseList))
	}
	if r, ok := node.(*parse.RangeNode); ok && r != nil {
		appendResult(processNode(root, parent, index, r.List))
		appendResult(processNode(root, parent, index, r.ElseList))
	}

	return ts, err
}

func processActionNode(root *templateFile, parent *parse.ListNode, index int, node parse.Node, funcName string) error {
	actionNode := node.(*parse.ActionNode)
	cmd := actionNode.Pipe.Cmds[0]
	_, name, field := getActionArgs(cmd)

	if name == root.Name() {
		return posErr{pos: int(actionNode.Pos), message: "cyclic reference"}
	}

	// validate for view and partial
	if invalidFuncType(root.typ, funcName) {
		return posErr{pos: int(actionNode.Pos), message: fmt.Sprintf("%s not supported", funcName)}
	}

	var arg parse.Node = &parse.DotNode{}

	// only handle if the function name is render or partial
	switch {
	case funcName == partialFunc.String():
		if field != nil {
			arg = field
		}
		if name == "" {
			return posErr{pos: int(actionNode.Pos), message: `path to partial file is not specified`}
		}
	case funcName == renderFunc.String():
		if name == "" {
			name = "body"
		}
	default:
		return nil
	}

	cmd.Args = []parse.Node{arg}
	actionNode.Pipe.Cmds = []*parse.CommandNode{cmd}

	tn := &parse.TemplateNode{
		NodeType: parse.NodeTemplate,
		Pos:      actionNode.Pos,
		Line:     actionNode.Line,
		Name:     name,
		Pipe:     actionNode.Pipe,
	}

	// replace the ActionNode with a TemplateNode.
	parent.Nodes[index] = tn
	return nil
}

func invalidFuncType(typ templateType, funcName string) bool {
	switch typ {
	case viewType:
		return funcName == renderFunc.String()
	case partialType:
		return funcName == renderFunc.String() || funcName == partialFunc.String()
	}

	return false
}

func getActionArgs(cmd *parse.CommandNode) (fn, file string, field *parse.FieldNode) {
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
	if len(cmd.Args) > 2 {
		if f, ok := cmd.Args[2].(*parse.FieldNode); ok {
			field = f
		}
	}
	return
}

// posErr tracks the position in the template file when a parse error occurs.
type posErr struct {
	pos     int
	message string
}

func (p posErr) Error() string {
	return p.message
}

func pos(body string, pos int) (line int, col int) {
	line = 1
	col = 1
	for i, char := range body {
		if i >= pos {
			break
		}

		if char == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

type nestingFunc string

func (t nestingFunc) String() string { return string(t) }

const (
	renderFunc  nestingFunc = "render"
	partialFunc nestingFunc = "partial"
)

type nestedFile struct {
	name string
	typ  nestingFunc
}
