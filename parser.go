package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type Tree struct {
	name  string
	input string
	lex   *lexer
	Root  []Node

	registeredTags map[string]Tag

	items     [5]Item
	peekCount int
}

// A Node is an element in the parse tree. The interface is trivial.
type Node interface {
	Position() int // byte position of start of node in full original input string
	Execute(ctx context.Context) (string, error)
}

// Base implements the Node interface
type Base struct {
	Start int
}

func (b *Base) Position() int                               { return b.Start }
func (b *Base) Execute(ctx context.Context) (string, error) { return "", nil }

// TextValue defines a text entry, and should be included as-is in the resulting
// template
type TextValue struct {
	Base
	Text string
}

// StringValue represents a string in an expression (e.g. an if-statement or a variable)
type StringValue struct {
	Base
	Val string
}

// IntValue represents a integer in an expression (e.g. an if-statement or a variable)
type IntValue struct {
	Base
	Val int
}

// getNumber returns either a integer or a float, depending on the incoming value
func getNumber(t Item) (Node, error) {
	if strings.ContainsRune(t.Val, '.') {
		// it's a float
	}
	v, err := strconv.Atoi(t.Val)
	if err != nil {
		return nil, err
	}

	return &IntValue{Base: Base{Start: t.Pos}, Val: v}, nil
}

// Comparison defines a comparison between two values
type Comparison struct {
	Base
	Type string
}

// Identifier is a name that gets evaluated at runtime, like a variable name or function name
type Identifier struct {
	Base
	Name string
}

// NewTree creates a new parser tree
func NewTree(name string) *Tree {
	return &Tree{name: name, registeredTags: map[string]Tag{}}
}

func (t *Tree) Next() Item {
	var i Item
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		i = <-t.lex.items
		t.items[0] = i
	}
	return t.items[t.peekCount]
}

func (t *Tree) Peek() Item {
	if t.peekCount > 0 {
		return t.items[t.peekCount-1]
	}
	t.peekCount = 1
	t.items[0] = <-t.lex.items
	return t.items[0]
}

func (t *Tree) consume() {
	if t.peekCount > 0 {
		t.peekCount--
		return
	}
	<-t.lex.items
}

func (t *Tree) ConsumeUntil(it ItemType) {
	for token := t.Next(); token.Typ != ItemEOF &&
		token.Typ != it; token = t.Next() {
	}
}

func (t *Tree) backup(i Item) {
	t.items[t.peekCount] = i
	t.peekCount++
}

// Parse builds the AST based on input
func (t *Tree) Parse(input string) error {
	l := newLexer(t.name, input)
	t.lex = l
	t.input = input
	return t.parse()
}

func (t *Tree) parse() error {
	t.Root = []Node{}
	for t.Peek().Typ != ItemEOF {
		token := t.Next()
		switch token.Typ {
		case ItemText:
			n := &TextValue{Text: token.Val}
			n.Start = token.Pos
			t.Root = append(t.Root, n)
		case ItemTagStart:
			n, err := t.tag()
			if err != nil {
				return err
			}
			t.Root = append(t.Root, n)
		default:
			return t.errorf("expected text or tag, got %s", token)
		}
	}
	return nil
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (t *Tree) errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// tag parses a tag node. The initial opening brace has already been parsed
func (t *Tree) tag() (Node, error) {
	tagname := t.Next()

	switch tagname.Typ {
	case ItemBlock:
		return t.newBlockStmt()
	case ItemIf:
		return t.newIfStmt()
	case ItemIdentifier:
		return t.newTag(tagname)
	}

	return nil, t.errorf("unknown tag %s", tagname.Val)
}

type Walker func(Node) Walker

func walk(fn Walker, nodeList []Node) (err error) {
	for k := range nodeList {
		sub := fn(nodeList[k])
		if sub == nil {
			continue
		}

		switch nodeList[k].(type) {
		case *BlockStmt:
			blk := nodeList[k].(*BlockStmt)
			err = walk(sub, blk.Body)
			if err != nil {
				return err
			}
		case *IfStmt:
			s := nodeList[k].(*IfStmt)
			err = walk(sub, s.Expression)
			if err != nil {
				return err
			}

			err = walk(sub, s.Body)
			if err != nil {
				return err
			}

			if s.Else != nil {
				err = walk(sub, []Node{s.Else})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *Tree) Walk(fn Walker) error {
	return walk(fn, t.Root)
}
