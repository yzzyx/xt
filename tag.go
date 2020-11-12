package main

// Tag defines the interface that all tags must fulfill
type Tag interface {
	Parse(s Stepper) (Node, error)
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
		return nil, t.Errorf("unknown tag %s", tagname.Val)
	}
	return tag.Parse(t)
}
