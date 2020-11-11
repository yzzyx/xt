package main

import (
	"fmt"

	"github.com/yzzyx/xt/lex"
)

// Tag defines the interface that all tags must fulfill
type Tag interface {
	Parse(s lex.Stepper) (Node, error)
}

// RegisterTag adds a new tag to the tree
func (t *Tree) RegisterTag(name string, tag Tag) {
	t.registeredTags[name] = tag
}

// newTag creates a node from a tag
func (t *Tree) newTag(tagname Item) (n Node, err error) {

	// Search for tag
	tag, ok := t.registeredTags[tagname.Val]
	if !ok {
		return nil, fmt.Errorf("unknown tag %s", tagname.Val)
	}
	return tag.Parse(t)
}
