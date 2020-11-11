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
	blockName := t.Next()
	if blockName.Typ != ItemIdentifier {
		return nil, t.errorf("expected identifier, got %s", blockName)
	}

	if t.Next().Typ != ItemTagEnd {
		return nil, t.errorf("expected end tag, got %s", t.Peek())
	}

	// now parse the contents of block
	body := []Node{}
Loop:
	for token := t.Next(); token.Typ != ItemEOF; token = t.Next() {
		switch token.Typ {
		case ItemText:
			n = &TextValue{Base: Base{Start: token.Pos}, Text: token.Val}
		case ItemTagStart:
			tagname := t.Peek()
			if tagname.Typ == ItemIdentifier &&
				tagname.Val == "endblock" {
				t.ConsumeUntil(ItemTagEnd)
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
		Base: Base{Start: blockName.Pos},
		Name: blockName.Val,
		Body: body,
	}

	return block, nil
}
