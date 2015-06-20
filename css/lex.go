package css

import (
	"strings"
	"unicode/utf8"
)

//see http://cuddle.googlecode.com/hg/talk/lex.html

type itemType int
type item struct {
	typ   itemType
	value string
}

func (i item) String() string {
	str := i.typ.String() + "\n"
	str += i.value
	return str
}

type lexer struct {
	input      string
	start      int
	current    int
	width      int
	stateStack []stateFunction
	items      chan *item
}
type stateFunction func(*lexer) stateFunction

// golang.org/x/tools/cmd/stringer
//go:generate stringer -type=itemType
const (
	itemError itemType = iota
	itemRoot
	itemSelector
	itemComment
	itemLeftBrace    // '{'
	itemRightBrace   // '}'
	itemCommentStart // '/*'
	itemCommentEnd   // '*/'
	itemProperty     //
	itemValue        //
	itemSeparator    // ':'
	itemTerminator   // ';'
)

const (
	leftBrace    = "{"
	rightBrace   = "}"
	commentStart = "/*"
	commentEnd   = "*/"
	separator    = ":"
	terminator   = ";"
	eof          = -1
)

var lexLeftBrace stateFunction
var lexSeparator stateFunction
var lexTerminator stateFunction
var lexRightBrace stateFunction
var lexCommentStart stateFunction

func init() {
	lexLeftBrace = lexConst(leftBrace, itemLeftBrace, lexProperty)
	lexSeparator = lexConst(separator, itemSeparator, lexValue)
	lexTerminator = lexConst(terminator, itemTerminator, lexProperty)
	lexRightBrace = lexConst(rightBrace, itemRightBrace, lexSelector)
	lexCommentStart = lexConst(commentStart, itemCommentStart, lexComment)
}

func (lex *lexer) run() {
	lex.items <- &item{
		itemRoot,
		"",
	}
	for state := lexSelector; state != nil; {
		state = state(lex)
	}
	close(lex.items)
}

func (lex *lexer) emit(it itemType) {
	value := strings.TrimSpace(lex.input[lex.start:lex.current])
	lex.ignore()

	if len(value) == 0 {
		return
	}

	lex.items <- &item{
		it,
		value,
	}
}

func (lex *lexer) next() (char rune) {
	if lex.current >= len(lex.input) {
		lex.width = 0
		return eof
	}
	char, lex.width = utf8.DecodeRuneInString(lex.input[lex.current:])
	lex.current += lex.width
	return char
}

func (lex *lexer) peek(prefix string) bool {
	return strings.HasPrefix(lex.input[lex.current:], prefix)
}

func (lex *lexer) ignore() {
	lex.start = lex.current
}

func lexConst(token string, it itemType, nextState stateFunction) stateFunction {
	return func(lex *lexer) stateFunction {
		lex.current += len(token)
		lex.emit(it)
		return nextState
	}
}

func lexSelector(lex *lexer) stateFunction {
	for {
		if lex.peek(leftBrace) {
			lex.emit(itemSelector)
			return lexLeftBrace
		} else if lex.peek(commentStart) {
			lex.emit(itemSelector)
			lex.stateStack = append(lex.stateStack, lexSelector)
			return lexCommentStart
		}
		if lex.next() == eof {
			break
		}
	}

	return nil
}

func lexProperty(lex *lexer) stateFunction {
	for {
		if lex.peek(rightBrace) {
			lex.emit(itemProperty)
			return lexRightBrace
		} else if lex.peek(commentStart) {
			lex.emit(itemProperty)
			lex.stateStack = append(lex.stateStack, lexProperty)
			return lexCommentStart
		} else if lex.peek(separator) {
			lex.emit(itemProperty)
			return lexSeparator
		}

		if lex.next() == eof {
			break
		}
	}

	return nil
}

func lexValue(lex *lexer) stateFunction {
	for {
		if lex.peek(terminator) {
			lex.emit(itemValue)
			return lexTerminator
		}

		if lex.next() == eof {
			break
		}
	}

	return nil
}

func lexCommentEnd(lex *lexer) stateFunction {
	lex.current += len(commentEnd)
	lex.emit(itemCommentEnd)

	nextState := lex.stateStack[len(lex.stateStack)-1]
	lex.stateStack = lex.stateStack[:len(lex.stateStack)-1]
	return nextState
}

func lexComment(lex *lexer) stateFunction {
	for {
		if lex.peek(commentEnd) {
			lex.emit(itemComment)
			return lexCommentEnd
		}
		if lex.next() == eof {
			break
		}
	}

	return nil
}
