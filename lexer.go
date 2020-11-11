package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Item defines a single entity in a template
type Item struct {
	Typ  ItemType // The type of this item.
	Pos  Pos      // The starting position, in bytes, of this item in the input string.
	Val  string   // The value of this item.
	Line int      // The line number at the start of this item.
	Col  int      // Column of the current item
}

// String returns the basic string representation of an item
func (i Item) String() string {
	return fmt.Sprintf("%02d:%02d %s - %s", i.Line, i.Pos, i.Typ, i.Val)
}

// ItemType identifies the type of lex items.
type ItemType int

// Pos is used to describe the position of an item
type Pos int

// Below is the definition of all the base item types available
const (
	ItemError      ItemType = iota // error occurred; value is text of error
	ItemBool                       // boolean constant
	ItemChar                       // printable ASCII character; grab bag for comma etc.
	ItemAssign                     // equals ('=') introducing an assignment
	ItemComparison                 // comparison '==', '>', '>=', '<', '<=', '!='
	ItemEOF
	ItemField      // alphanumeric identifier starting with '.'
	ItemIdentifier // alphanumeric identifier not starting with '.'
	ItemTagStart   // left action delimiter
	ItemLeftParen  // '(' inside action
	ItemNumber     // simple number, including imaginary
	ItemPipe       // pipe symbol
	ItemTagEnd     // right action delimiter
	ItemRightParen // ')' inside action
	ItemSpace      // run of spaces separating arguments
	ItemString     // quoted string (includes quotes)
	ItemText       // plain text
	ItemVariable   // variable starting with '$', such as '$' or  '$1' or '$hello'
	ItemVarStart   // Start of a variable '{{'
	ItemVarEnd     // End of a variable '}}'
	// Keywords appear after all the rest.
	ItemKeyword // used only to delimit the keywords
	ItemBlock   // block keyword
	ItemElse    // else keyword
	ItemElIf    // elif keyword
	ItemEnd     // end keyword
	ItemIf      // if keyword
)

var itemTypeMap = map[ItemType]string{
	ItemError:      "error",
	ItemBool:       "bool",
	ItemChar:       "char",
	ItemComparison: "comparison",
	ItemAssign:     "assign",
	ItemEOF:        "EOF",
	ItemIdentifier: "identifier",
	ItemTagStart:   "left-delim",
	ItemLeftParen:  "left-paren",
	ItemNumber:     "number",
	ItemPipe:       "pipe",
	ItemTagEnd:     "right-delim",
	ItemRightParen: "right-paren",
	ItemSpace:      "space",
	ItemString:     "string",
	ItemText:       "text",
	ItemVariable:   "variable",

	ItemBlock: "block",
	ItemElse:  "else",
	ItemElIf:  "elif",
	ItemEnd:   "end",
	ItemIf:    "if",
}

// String returns the printable name of an item type
func (i ItemType) String() string {
	return itemTypeMap[i]
}

const (
	delimTagStart = "{%"
	delimTagEnd   = "%}"
	delimVarStart = "{{"
	delimVarEnd   = "}}"
)

const eof = -1

type lexer struct {
	name       string
	line       int
	startLine  int
	col        int
	input      string
	parenDepth int

	pos   Pos       // current position in the input
	start Pos       // start position of this item
	width Pos       // width of last rune read from input
	items chan Item // channel of scanned items
}

type stateFn func(*lexer) stateFn

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t ItemType) {
	l.items <- Item{
		Typ:  t,
		Pos:  l.pos,
		Val:  l.input[l.start:l.pos],
		Line: l.startLine,
		Col:  l.col,
	}
	l.start = l.pos
	l.startLine = l.line
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{
		Typ:  ItemError,
		Pos:  l.start,
		Val:  fmt.Sprintf(format, args...),
		Line: l.startLine,
		Col:  l.col,
	}
	return nil
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	//if left == "" {
	//	left = delimTagStart
	//}
	//if right == "" {
	//	right = delimTagEnd
	//}
	l := &lexer{
		name:      name,
		input:     input,
		items:     make(chan Item),
		line:      1,
		startLine: 1,
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

// lexText scans until an opening action delimiter, "{%".
func lexText(l *lexer) stateFn {
	l.width = 0

	//if x := strings.Index(l.input[l.pos:], l.delimTagStart); x >= 0 {
	if x := strings.IndexRune(l.input[l.pos:], '{'); x >= 0 {
		var nextFunc stateFn
		l.pos += Pos(x)

		if strings.HasPrefix(l.input[l.pos:], delimTagStart) {
			nextFunc = lexTagStart
		} else if strings.HasPrefix(l.input[l.pos:], delimVarStart) {
			nextFunc = lexVarStart
		} else {
			return l.errorf("unexpected sequence %s, expected {%% or {{", l.input[l.pos:l.pos+1])
		}

		if l.pos > l.start {
			l.line += strings.Count(l.input[l.start:l.pos], "\n")
			l.emit(ItemText)
		}
		l.ignore()
		return nextFunc

	}
	l.pos = Pos(len(l.input))
	// Correctly reached EOF.
	if l.pos > l.start {
		l.line += strings.Count(l.input[l.start:l.pos], "\n")
		l.emit(ItemText)
	}
	l.emit(ItemEOF)
	return nil
}

// lexTagStart scans the start tag marker '{%'
func lexTagStart(l *lexer) stateFn {
	l.pos += Pos(len(delimTagStart))
	l.emit(ItemTagStart)
	return lexInsideTag
}

// lexTagEnd scans the end tag marker '%}'
func lexTagEnd(l *lexer) stateFn {
	l.pos += Pos(len(delimTagEnd))
	l.emit(ItemTagEnd)
	return lexText
}

// lexVarStart is the start of a variable '{{'
func lexVarStart(l *lexer) stateFn {
	l.pos += Pos(len(delimVarStart))
	l.emit(ItemVarStart)
	return lexInsideTag
}

// lexVarEnd is the start of a variable '}}'
func lexVarEnd(l *lexer) stateFn {
	l.pos += Pos(len(delimVarEnd))
	l.emit(ItemVarEnd)
	return lexText
}

// lexInsideTag scans the elements inside action delimiters.
func lexInsideTag(l *lexer) stateFn {
	// Either number, quoted string, or identifier.
	// Spaces separate arguments; runs of spaces turn into itemSpace.
	// Pipe symbols separate and are emitted.
	if strings.HasPrefix(l.input[l.pos:], delimTagEnd) { // Without trim marker.
		if l.parenDepth > 0 {
			return l.errorf("missing right paren")
		}
		return lexTagEnd
	} else if strings.HasPrefix(l.input[l.pos:], delimVarEnd) { // Without trim marker.
		if l.parenDepth > 0 {
			return l.errorf("missing right paren")
		}
		return lexVarEnd
	}

	switch r := l.next(); {
	case r == eof || isEndOfLine(r):
		return l.errorf("unclosed action")
	case isSpace(r):
		l.ignore()
	case r == '!':
		rn := l.next()
		if rn != '=' {
			return l.errorf("expected = after !")
		}
		l.emit(ItemComparison)
	case r == '>' || r == '<':
		rn := l.next()
		if rn != '=' {
			l.backup()
		}
		l.emit(ItemComparison)
	case r == '=':
		rn := l.next()
		if rn == '=' {
			l.emit(ItemComparison)
		} else {
			l.backup()
			l.emit(ItemAssign)
		}
	case r == '|':
		l.emit(ItemPipe)
	case r == '"':
		return lexQuote
	case r == '\'':
		return lexSingleQuote
	case r == '+' || r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	case r == '(':
		l.emit(ItemLeftParen)
		l.parenDepth++
	case r == ')':
		l.emit(ItemRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren %#U", r)
		}
	case r <= unicode.MaxASCII && unicode.IsPrint(r):
		l.emit(ItemChar)
		return lexInsideTag
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
	return lexInsideTag
}

// lexNumber lexes a number
func lexNumber(l *lexer) stateFn {

	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789_"
	if l.accept("0") {
		// Note: Leading 0 does not mean octal in floats.
		if l.accept("xX") {
			digits = "0123456789abcdefABCDEF_"
		} else if l.accept("oO") {
			digits = "01234567_"
		} else if l.accept("bB") {
			digits = "01_"
		}
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if len(digits) == 10+1 && l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	if len(digits) == 16+6+1 && l.accept("pP") {
		l.accept("+-")
		l.acceptRun("0123456789_")
	}
	l.emit(ItemNumber)
	return lexInsideTag
}

// lexQuote lexes a quoted string and returns to parent function
func lexQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}
	l.emit(ItemString)
	return lexInsideTag
}

// lexQuote lexes a single-quoted string and returns to parent function
func lexSingleQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated quoted string")
		case '\'':
			break Loop
		}
	}
	l.emit(ItemString)
	return lexInsideTag
}

var typeMap = map[string]ItemType{
	"block": ItemBlock,
	"if":    ItemIf,
	"else":  ItemElse,
	"elif":  ItemElIf,
}

func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isSpace(r):
			l.backup()
			word := l.input[l.start:l.pos]
			switch {
			case typeMap[word] > ItemKeyword:
				l.emit(typeMap[word])
			case word[0] == '.':
				l.emit(ItemField)
			case word == "true", word == "false":
				l.emit(ItemBool)
			default:
				l.emit(ItemIdentifier)
			}
			break Loop
		case isAlphaNumeric(r):
			// absorb.
		default:
			return l.errorf("bad character %#U", r)
		}
	}
	return lexInsideTag
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
