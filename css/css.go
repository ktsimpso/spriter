package css

import (
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"
)

func GetPaths(filename string) ([]string, error) {
	cssFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer cssFile.Close()

	cssFileBytes, err := ioutil.ReadAll(cssFile)
	if err != nil {
		return nil, err
	}

	cssString := string(cssFileBytes[:])
	items := make(chan item)

	lex := &lexer{
		input:      cssString,
		stateStack: []stateFunction{},
		items:      items,
	}

	go lex.run()
	findingUrls := false
	value := ""
	urls := []string{}

	// TODO: make this more robust
	for item := range items {
		if !findingUrls && item.typ == itemProperty && (item.value == "background" || item.value == "background-image") {
			findingUrls = true
			continue
		}

		if findingUrls && item.typ == itemTerminator {
			n := strings.Index(value, "url(")
			url := value[n+4 : strings.Index(value[n:], ")")]

			if !strings.HasPrefix(url, "http") {
				urls = append(urls, url)
			}

			findingUrls = false
			value = ""
		}

		if findingUrls && item.typ == itemValue {
			value += item.value + " "
		}
	}

	return urls, nil
}

//see http://cuddle.googlecode.com/hg/talk/lex.html

type itemType int
type item struct {
	typ   itemType
	value string
}

func (i *item) String() string {
	str := ""
	switch i.typ {
	case itemSelector:
		str += "Selector\n"
	case itemComment:
		str += "Comment\n"
	case itemCommentStart:
		str += "CommentStart\n"
	case itemCommentEnd:
		str += "CommentEnd\n"
	case itemLeftBrace:
		str += "LeftBrace\n"
	case itemRightBrace:
		str += "RightBrace\n"
	case itemProperty:
		str += "Property\n"
	case itemTerminator:
		str += "Terminator\n"
	case itemSeparator:
		str += "Separator\n"
	case itemValue:
		str += "Value\n"
	default:
		str += "Unkown token\n"
	}

	str += i.value
	return str
}

type lexer struct {
	input      string
	start      int
	current    int
	width      int
	stateStack []stateFunction
	items      chan item
}
type stateFunction func(*lexer) stateFunction

const (
	itemError itemType = iota
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

func (lex *lexer) run() {
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

	lex.items <- item{
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

func lexLeftBrace(lex *lexer) stateFunction {
	lex.current += len(leftBrace)
	lex.emit(itemLeftBrace)
	return lexProperty
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

func lexSeparator(lex *lexer) stateFunction {
	lex.current += len(separator)
	lex.emit(itemSeparator)
	return lexValue
}

func lexValue(lex *lexer) stateFunction {
	for {
		if lex.peek(terminator) {
			lex.emit(itemValue)
			return lexTerminator
		} else if lex.peek(commentStart) {
			lex.emit(itemValue)
			lex.stateStack = append(lex.stateStack, lexValue)
			return lexCommentStart
		}

		if lex.next() == eof {
			break
		}
	}

	return nil
}

func lexTerminator(lex *lexer) stateFunction {
	lex.current += len(terminator)
	lex.emit(itemTerminator)
	return lexProperty
}

func lexRightBrace(lex *lexer) stateFunction {
	lex.current += len(rightBrace)
	lex.emit(itemRightBrace)
	return lexSelector
}

func lexCommentStart(lex *lexer) stateFunction {
	lex.current += len(commentStart)
	lex.emit(itemCommentStart)
	return lexComment
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
