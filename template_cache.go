package gomplate

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	gotemplate "text/template"
	"text/template/parse"
	"time"

	"github.com/patrickmn/go-cache"
)

// Cached templates own immutable parse trees and only the static functions they
// reference. Request functions are registered on execution-local clones, never
// retained in the cache. Static renders can execute the cached template directly.
var goTemplateCache = cache.New(time.Hour, time.Hour)

// goTemplateCacheKey identifies one effective parsing pass. Function names affect
// parsing (including whether break/continue are keywords), but their values are
// rebound for every render. Length prefixes keep arbitrary source/key bytes distinct.
func (t Template) goTemplateCacheKey() string {
	fields := [4]string{t.CacheKey, t.LeftDelim, t.RightDelim, t.Template}
	var size [binary.MaxVarintLen64]byte
	capacity := 0
	for _, field := range fields {
		capacity += len(field) + binary.PutUvarint(size[:], uint64(len(field)))
	}
	names := make([]string, 0, len(t.Functions))
	for name := range t.Functions {
		names = append(names, name)
		capacity += len(name) + binary.PutUvarint(size[:], uint64(len(name)))
	}
	sort.Strings(names)

	var key strings.Builder
	key.Grow(capacity)
	write := func(value string) {
		n := binary.PutUvarint(size[:], uint64(len(value)))
		key.Write(size[:n])
		key.WriteString(value)
	}
	for _, field := range fields {
		write(field)
	}
	for _, name := range names {
		write(name)
	}
	return key.String()
}

// reusableGoTemplate rebuilds an execution-only template from parsed trees.
// Go releases parsing function maps after Parse; AddParseTree shares only the
// immutable syntax. Rebuilding avoids both captured closures and copying the
// entire built-in registry on every dynamic cache hit.
func reusableGoTemplate(parsed *gotemplate.Template) (*gotemplate.Template, error) {
	compiled := gotemplate.New(parsed.Name())
	functions := make(gotemplate.FuncMap)
	for _, named := range parsed.Templates() {
		if err := collectGoTemplateFunctions(named.Root, functions); err != nil {
			return nil, err
		}
		if _, err := compiled.AddParseTree(named.Name(), named.Tree); err != nil {
			return nil, err
		}
	}
	return compiled.Funcs(functions), nil
}

func collectGoTemplateFunctions(node parse.Node, functions gotemplate.FuncMap) error {
	var children []parse.Node
	switch node := node.(type) {
	case *parse.IdentifierNode:
		if function, ok := funcMap[node.Ident]; ok {
			functions[node.Ident] = function
		}
	case *parse.ListNode:
		if node != nil {
			children = node.Nodes
		}
	case *parse.ActionNode:
		children = []parse.Node{node.Pipe}
	case *parse.PipeNode:
		if node != nil {
			for _, command := range node.Cmds {
				if err := collectGoTemplateFunctions(command, functions); err != nil {
					return err
				}
			}
		}
	case *parse.CommandNode:
		children = node.Args
	case *parse.ChainNode:
		children = []parse.Node{node.Node}
	case *parse.IfNode:
		children = []parse.Node{node.Pipe, node.List, node.ElseList}
	case *parse.RangeNode:
		children = []parse.Node{node.Pipe, node.List, node.ElseList}
	case *parse.WithNode:
		children = []parse.Node{node.Pipe, node.List, node.ElseList}
	case *parse.TemplateNode:
		children = []parse.Node{node.Pipe}
	case nil, *parse.TextNode, *parse.CommentNode, *parse.BoolNode, *parse.DotNode,
		*parse.NilNode, *parse.NumberNode, *parse.StringNode, *parse.FieldNode,
		*parse.VariableNode, *parse.BreakNode, *parse.ContinueNode:
	default:
		return fmt.Errorf("unsupported Go template node %T", node)
	}
	for _, child := range children {
		if err := collectGoTemplateFunctions(child, functions); err != nil {
			return err
		}
	}
	return nil
}
