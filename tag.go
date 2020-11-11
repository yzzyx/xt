package main

import (
	"fmt"
)

// Tag defines the interface that all tags must fulfill
type Tag interface {
	Parse(t *Tree, p Pos) (Node, error)
}

// RegisterTag adds a new tag to the tree
func (t *Tree) RegisterTag(name string, tag Tag) {
	t.registeredTags[name] = tag
}

// newTag creates a node from a tag
func (t *Tree) newTag(tagname item) (n Node, err error) {

	// Search for tag
	tag, ok := t.registeredTags[tagname.val]
	if !ok {
		return nil, fmt.Errorf("unknown tag %s", tagname.val)
	}
	return tag.Parse(t, tagname.pos)
}
