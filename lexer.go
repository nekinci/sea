package main

import (
	"fmt"
	"unicode"
)

type Lexer struct {
	filename                      string
	input                         string
	inputLen                      int
	curTok                        Token
	curVal                        string
	comments                      bool
	col, line, pos                int
	startCol, startLine, startPos int
	lastLine                      int
}

// Start returns current token's start position
func (l *Lexer) Start() Pos {
	return Pos{
		Col:    l.startCol,
		Line:   l.startLine,
		Offset: l.startPos,
	}
}

// End returns current token's end position
func (l *Lexer) End() Pos {
	return Pos{
		Col:    l.col,
		Line:   l.line,
		Offset: l.pos,
	}
}

func (l *Lexer) isNewLine() bool {
	if l.input[l.pos] == '\n' {
		return true
	}

	return false
}

func (l *Lexer) operatorToken() Token {
	ch := l.input[l.pos]
	l.pos = l.pos + 1

	switch ch {
	case '+':
		if l.input[l.pos] == '+' {
			// TODO
		}

		if l.input[l.pos] == '=' {
			// TODO
		}

		return TokPlus
	case '-':
		if l.input[l.pos] == '-' {
			// TODO
		}

		if l.input[l.pos] == '=' {
			// TODO
		}

		return TokMinus
	case '*':
		if l.input[l.pos] == '=' {
		}
		return TokMultiply
	case '/':
		if l.input[l.pos] == '=' {
		}
		return TokDivision
	case '%':
		if l.input[l.pos] == '=' {
		}
		return TokMod
	}

	panic("unreachable")

}

func (l *Lexer) printTokens() {
	next, s := l.next()
	for next != EOF {
		fmt.Printf("Token[%s - %s]\n", next, s)
		next, s = l.next()
	}

}

func (l *Lexer) backup(len int) {
	l.pos = l.pos - len
	l.col = l.col - len
	// TODO line decreasing
	Assert(l.pos >= 0 && l.pos < l.inputLen, "lexer position out of bounds")
}

func (l *Lexer) nextAndBackup() (Token, string) {
	token, v := l.next()
	l.backup(len(v))
	return token, v
}

func (l *Lexer) skipWhitespace() {
	for l.pos < l.inputLen && isWhitespace(l.input[l.pos]) {
		if l.isNewLine() {
			l.line++
			l.col = -1 // To adjust to the 0 position after the if statement
		}
		l.pos = l.pos + 1
		l.col++
	}
}

func (l *Lexer) skipLineComment() Token {
	for !l.isNewLine() {
		l.pos = l.pos + 1
		l.col = l.col + 1
	}
	l.line = l.line + 1
	l.pos = l.pos + 1
	l.col = 0
	return TokSingleComment
}

func (l *Lexer) tok() Token {
	if l.pos >= len(l.input) {
		return EOF
	}

	l.startPos = l.pos
	l.startLine = l.line
	l.startCol = l.col
	c := l.input[l.pos]
	switch {
	case c == '#':
		l.skipLineComment()
		return l.tok()
	case isWhitespace(c):
		l.skipWhitespace()
		return l.tok()
	case isDigit(c):
		// TODO handle floating point numbers, negative numbers or e signed integers
		for c >= '0' && c <= '9' {
			l.pos++
			l.col++
			c = l.input[l.pos]
		}
		return TokNumber
	case c == '+' || c == '-' || c == '*' || c == '/' || c == '%':
		if c == '/' && l.inputLen > l.pos+1 && l.input[l.pos+1] == '*' {
			l.pos += 2
			l.col += 2
			for l.pos+1 < l.inputLen && !(l.input[l.pos] == '*' && l.input[l.pos+1] == '/') {
				if l.input[l.pos] == '\n' {
					l.col = -1
					l.line++
				}
				l.pos++
				l.col++
			}
			l.pos += 2
			l.col += 2
			return l.tok()
		}
		return l.operatorToken()
	case c == '{':
		l.pos++
		l.col++
		return TokLBrace
	case c == '}':
		l.pos++
		l.col++
		return TokRBrace
	case c == '(':
		l.pos++
		l.col++
		return TokLParen
	case c == ')':
		l.pos++
		l.col++
		return TokRParen
	case c == ',':
		l.pos++
		l.col++
		return TokComma
	case c == ';':
		l.pos++
		l.col++
		return TokSemicolon
	case c == ':':
		l.pos++
		l.col++
		return TokColon
	case unicode.IsLetter(rune(c)):

		for l.pos < l.inputLen && (unicode.IsDigit(rune(l.input[l.pos])) || unicode.IsLetter(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
			l.pos += 1
			l.col++
		}

		identifier := l.value()
		switch identifier {
		case "var":
			return TokVar
		case "fun":
			return TokFun
		case "return":
			return TokReturn
		case "true":
			return TokTrue
		case "false":
			return TokFalse
		case "if":
			return TokIf
		case "else":
			return TokElse
		case "for":
			return TokFor
		case "break":
			return TokBreak
		case "continue":
			return TokContinue
		case "extern":
			return TokExtern
		case "struct":
			return TokStruct
		case "impl":
			return TokImpl
		case "nil":
			return TokNil
		case "sizeof":
			return TokSizeof
		}

		return TokIdentifier
	case isDigit(c):
		// TODO . , e handle
		for l.pos < l.inputLen && isDigit(l.input[l.pos]) {
			l.pos += 1
			l.col++
		}
		return TokNumber
	case c == '"':
		first := true
		for l.pos < l.inputLen && (l.input[l.pos] != '"' || first) {
			l.pos += 1
			l.col++
			first = false
		}
		l.pos++
		l.col++
		return TokString
	case c == '=':
		l.pos += 1
		l.col++
		if l.input[l.pos] == '=' {
			l.pos += 1
			l.col++
			return TokEqual
		}

		return TokAssign
	case c == '!':
		l.pos++
		l.col++
		if l.input[l.pos] == '=' {
			l.pos += 1
			l.col++
			return TokNEqual
		}
		return TokNot
	case c == '&':
		l.pos++
		l.col++
		if l.input[l.pos] == '&' {
			l.pos += 1
			l.col++
			return TokAnd
		}
		return TokBAnd
	case c == '|':
		l.pos++
		l.col++
		if l.input[l.pos] == '|' {
			l.pos += 1
			l.col++
			return TokOr
		}
		return TokBOr
	case c == '^':
		l.pos++
		l.col++
		return TokXor
	case c == '<':
		l.pos++
		l.col++
		if l.input[l.pos] == '=' {
			l.pos += 1
			l.col++
			return TokLte
		}

		if l.input[l.pos] == '<' {
			l.pos++
			l.col++
			return TokLShift
		}

		return TokLt
	case c == '>':
		l.pos++
		l.col++
		if l.input[l.pos] == '=' {
			l.pos += 1
			l.col++
			return TokGte
		}

		if l.input[l.pos] == '>' {
			l.pos++
			l.col++
			return TokRShift
		}

		return TokGt
	case c == '[':
		l.pos++
		l.col++
		return TokLBracket
	case c == ']':
		l.pos++
		l.col++
		return TokRBracket
	case c == '.':
		l.pos++
		l.col++
		return TokDot

	}

	panic("unreachable: " + string(rune(c)))
}

func (l *Lexer) next() (Token, string) {

	l.lastLine = l.line
	tok := l.tok()
	l.curTok = tok
	l.curVal = l.value()

	return tok, l.curVal

}

func (l *Lexer) value() string {
	return l.input[l.startPos:l.pos]
}
