package main

import (
	"context"
	"strings"
)

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

// ExecuteNodes traverses a set of nodes and returns its combined strings
func ExecuteNodes(ctx context.Context, nodes []Node) (string, error) {
	b := strings.Builder{}
	for idx := range nodes {
		v, err := nodes[idx].Execute(ctx)
		if err != nil {
			return "", err
		}
		b.WriteString(v)
	}
	return b.String(), nil
}
