package main

// IfStmt defines an if-statement
// If expression is met, 'Body' should be executed.
// If not, Else should be executed
type IfStmt struct {
	Base
	Expression []Node
	Body       []Node
	Else       Node
}

// if statement:
//  {% if expression %}
//  [{% elif expression %}]
//  [{% else %}]
//  {% endif %}
func (t *Tree) newIfStmt() (n Node, err error) {
	start := t.items[0]
	expression := []Node{}
	for token := t.next(); token.Typ != ItemTagEnd; token = t.next() {
		if token.Typ == ItemEOF {
			return nil, t.errorf("expected end of tag, got EOF")
		}
		switch token.Typ {
		case ItemString:
			n = &StringValue{Base: Base{Start: token.Pos}, Val: token.Val}
		case ItemNumber:
			n, err = getNumber(token)
			if err != nil {
				return nil, err
			}
		case ItemComparison:
			n = &Comparison{Base: Base{Start: token.Pos}, Type: token.Val}
		//case itemLeftParen:
		//case itemRightParen:
		case ItemIdentifier:
			n = &Identifier{Base: Base{Start: token.Pos}, Name: token.Val}
		default:
			return nil, t.errorf("unexpected token in expression: %s", token)
		}
		expression = append(expression, n)
	}

	// now parse the contents of the if-stmt
	var token Item
	body := []Node{}
	var elseIfNode Node
	var elseNode Node
Loop:
	for token = t.next(); token.Typ != ItemEOF; token = t.next() {
		switch token.Typ {
		case ItemText:
			n = &TextValue{Base: Base{Start: token.Pos}, Text: token.Val}
		case ItemTagStart:
			tagname := t.peek()
			if tagname.Typ == ItemElIf {
				// Treat ElIf as a if-statement inside the 'else'-statement,
				// so we save it, and check if we have an actual else-stmt
				elseIfNode, err = t.newIfStmt()
				if err != nil {
					return nil, err
				}

				// bump token and tagname back on the stack,
				// in order for elseif-handling to work properly
				t.backup(Item{Typ: ItemTagEnd})
				t.backup(Item{Typ: ItemIdentifier, Val: "endif"})
				t.backup(token)
				continue
			} else if tagname.Typ == ItemElse {
				// Create an else body
				elseNode, err = t.newElseStmt()
				if err != nil {
					return nil, err
				}

				// bump token and tagname back on the stack,
				// in order for elseif-handling to work properly
				t.backup(Item{Typ: ItemTagEnd})
				t.backup(Item{Typ: ItemIdentifier, Val: "endif"})
				t.backup(token)
				continue
			}

			// If we're at endif, stop parsing
			if tagname.Typ == ItemIdentifier &&
				tagname.Val == "endif" {
				t.consumeUntil(ItemTagEnd)
				break Loop
			}

			n, err = t.tag()
			if err != nil {
				return nil, err
			}
		}
		body = append(body, n)
	}

	if token.Typ == ItemEOF {
		return nil, t.errorf("expected 'endif'-tag, got end-of-file")
	}

	// convert the following pattern
	//   {% if abc %}
	//   {% elif def %}
	//   {% endif %}
	// to
	//   {% if abc %}
	//   {% else %}
	//     {% if def %}
	//     {% endif %}
	//   {% endif %}
	if elseIfNode != nil {
		elseBody := []Node{elseIfNode}
		if elseNode != nil {
			elseBody = append(elseBody, elseNode)
		}

		elseNode = &BlockStmt{
			Base:      Base{Start: elseIfNode.Position()},
			Name:      "",
			Arguments: nil,
			Body:      elseBody,
		}
	}

	block := &IfStmt{
		Base:       Base{Start: start.Pos},
		Expression: expression,
		Body:       body,
		Else:       elseNode,
	}

	return block, nil
}

// else statement:
//  {% else %}
//    ...
//  {% endif %}
func (t *Tree) newElseStmt() (n Node, err error) {
	start := t.next()
	token := t.next()
	if token.Typ != ItemTagEnd {
		return nil, t.errorf("unexpected extra arguments to 'else' statement: %s", token)
	}

	body := []Node{}
Loop:
	for token := t.next(); token.Typ != ItemEOF; token = t.next() {
		switch token.Typ {
		case ItemText:
			n = &TextValue{Base: Base{Start: token.Pos}, Text: token.Val}
		case ItemTagStart:
			tagname := t.peek()
			if tagname.Typ == ItemIdentifier &&
				tagname.Val == "endif" {
				t.consumeUntil(ItemTagEnd)
				break Loop
			}

			n, err = t.tag()
			if err != nil {
				return nil, err
			}
		}
		body = append(body, n)
	}

	stmt := &BlockStmt{
		Base:      Base{Start: start.Pos},
		Name:      "",
		Arguments: nil,
		Body:      body,
	}
	return stmt, nil
}
