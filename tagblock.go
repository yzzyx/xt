package main

// BlockStmt defines a block in a template
// Unnamed blocks, with name set to "", can be used to
// wrap statements, e.g. in an else statement
type BlockStmt struct {
	Base
	Name      string
	Arguments []Node
	Body      []Node
}

// block statement:
//  {% block <name:identifier> [with...] %}
func (t *Tree) newBlockStmt() (n Node, err error) {
	blockName := t.next()
	if blockName.typ != itemIdentifier {
		return nil, t.errorf("expected identifier, got %s", blockName)
	}

	if t.next().typ != itemTagEnd {
		return nil, t.errorf("expected end tag, got %s", t.peek())
	}

	// now parse the contents of block
	body := []Node{}
Loop:
	for token := t.next(); token.typ != itemEOF; token = t.next() {
		switch token.typ {
		case itemText:
			n = &TextValue{Base: Base{Start: token.pos}, Text: token.val}
		case itemTagStart:
			tagname := t.peek()
			if tagname.typ == itemIdentifier &&
				tagname.val == "endblock" {
				t.consumeUntil(itemTagEnd)
				break Loop
			}

			n, err = t.tag()
			if err != nil {
				return nil, err
			}
		}
		body = append(body, n)
	}

	block := &BlockStmt{
		Base: Base{Start: blockName.pos},
		Name: blockName.val,
		Body: body,
	}

	return block, nil
}
