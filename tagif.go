package main

// IfStmt defines an if-statement
// If expression is met, 'Body' should be executed.
// If not, Else should be executed
type IfStmt struct {
	Start      Pos
	Expression []Node
	Body       []Node
	Else       Node
}

// Position returns the start position of the statement
func (s *IfStmt) Position() Pos { return s.Start }

// if statement:
//  {% if expression %}
//  [{% elif expression %}]
//  [{% else %}]
//  {% endif %}
func (t *Tree) newIfStmt() (n Node, err error) {
	start := t.items[0]
	expression := []Node{}
	for token := t.next(); token.typ != itemTagEnd; token = t.next() {
		if token.typ == itemEOF {
			return nil, t.errorf("expected end of tag, got EOF")
		}
		switch token.typ {
		case itemString:
			n = &StringValue{Start: token.pos, Val: token.val}
		//case itemNumber:
		//case itemComparison:
		//case itemLeftParen:
		//case itemRightParen:
		case itemIdentifier:
			n = &Identifier{Start: token.pos, Name: token.val}
		default:
			return nil, t.errorf("unexpected token in expression: %s", token)
		}
		expression = append(expression, n)
	}

	// now parse the contents of the if-stmt
	var token item
	body := []Node{}
	var elseIfNode Node
	var elseNode Node
Loop:
	for token = t.next(); token.typ != itemEOF; token = t.next() {
		switch token.typ {
		case itemText:
			n = &TextValue{Start: token.pos, Text: token.val}
		case itemTagStart:
			tagname := t.peek()
			if tagname.typ == itemElIf {
				// Treat ElIf as a if-statement inside the 'else'-statement,
				// so we save it, and check if we have an actual else-stmt
				elseIfNode, err = t.newIfStmt()
				if err != nil {
					return nil, err
				}

				// bump token and tagname back on the stack,
				// in order for elseif-handling to work properly
				t.backup(item{typ: itemTagEnd})
				t.backup(item{typ: itemIdentifier, val: "endif"})
				t.backup(token)
				continue
			} else if tagname.typ == itemElse {
				// Create an else body
				elseNode, err = t.newElseStmt()
				if err != nil {
					return nil, err
				}

				// bump token and tagname back on the stack,
				// in order for elseif-handling to work properly
				t.backup(item{typ: itemTagEnd})
				t.backup(item{typ: itemIdentifier, val: "endif"})
				t.backup(token)
				continue
			}

			// If we're at endif, stop parsing
			if tagname.typ == itemIdentifier &&
				tagname.val == "endif" {
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

	if token.typ == itemEOF {
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
			Start:     elseIfNode.Position(),
			Name:      "",
			Arguments: nil,
			Body:      elseBody,
		}
	}

	block := &IfStmt{
		Start:      start.pos,
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
	if token.typ != itemTagEnd {
		return nil, t.errorf("unexpected extra arguments to 'else' statement: %s", token)
	}

	body := []Node{}
Loop:
	for token := t.next(); token.typ != itemEOF; token = t.next() {
		switch token.typ {
		case itemText:
			n = &TextValue{Start: token.pos, Text: token.val}
		case itemTagStart:
			tagname := t.peek()
			if tagname.typ == itemIdentifier &&
				tagname.val == "endif" {
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

	stmt := &BlockStmt{
		Start:     start.pos,
		Name:      "",
		Arguments: nil,
		Body:      body,
	}
	return stmt, nil
}
