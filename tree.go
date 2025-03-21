package mold

import "text/template/parse"

// processTree traverses the node tree and swaps render and partial declarations with equivalent template calls.
// It returns all referenced templates encountered during the traversal.
func processTree(parent *parse.ListNode, index int, node parse.Node, render, partial bool) []string {
	var ts []string
	if a, ok := node.(*parse.ActionNode); ok {
		if len(a.Pipe.Cmds) > 0 {
			funcName, tname, _ := getActionArgs(a.Pipe.Cmds[0])
			if funcName == renderFunc || funcName == partialFunc {
				processActionNode(parent, index, node, render, partial)
			}
			if tname != "" {
				ts = append(ts, tname)
			}
		}
	}
	if l, ok := node.(*parse.ListNode); ok {
		for i, n := range l.Nodes {
			ts = append(ts, processTree(l, i, n, render, partial)...)
		}
	}
	if i, ok := node.(*parse.IfNode); ok {
		ts = append(ts, processTree(parent, index, i.List, render, partial)...)
		ts = append(ts, processTree(parent, index, i.ElseList, render, partial)...)
	}
	if r, ok := node.(*parse.RangeNode); ok {
		ts = append(ts, processTree(parent, index, r.List, render, partial)...)
		ts = append(ts, processTree(parent, index, r.ElseList, render, partial)...)
	}

	return ts
}

func processActionNode(parent *parse.ListNode, index int, node parse.Node, render, partial bool) {
	if parent == nil {
		// this should never happen
		panic("parent node is nil")
	}

	actionNode := node.(*parse.ActionNode)
	cmd := actionNode.Pipe.Cmds[0]
	funcName, name, field := getActionArgs(cmd)

	var arg parse.Node = &parse.DotNode{}

	// only handle if the function name is render or partial
	switch {
	case funcName == partialFunc && partial:
		if field != nil {
			arg = field
		}
	case funcName == renderFunc && render:
		if name == "" {
			name = "body"
		}
	default:
		return
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

	// replace the action node with a template node.
	parent.Nodes[index] = tn
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
