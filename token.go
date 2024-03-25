package main

type Token string

func (t Token) String() string {
	return string(t)
}

const (
	EOF              Token = "<eof>"
	TokPlus          Token = "+"
	TokIncr          Token = "++"
	TokMinus         Token = "-"
	TokDecr          Token = "--"
	TokMultiply      Token = "*"
	TokDivision      Token = "/"
	TokMod           Token = "%"
	TokEqual         Token = "=="
	TokIdentifier    Token = "<identifier>"
	TokNumber        Token = "<number_literal>"
	TokFloat         Token = "<float_literal>"
	TokString        Token = "<string_literal>"
	TokVar           Token = "var"
	TokAssign        Token = "="
	TokColon         Token = ":"
	TokSemicolon     Token = ";"
	TokDef           Token = "def"
	TokFun           Token = "fun"
	TokLParen        Token = "("
	TokRParen        Token = ")"
	TokComma         Token = ","
	TokLBrace        Token = "{"
	TokRBrace        Token = "}"
	TokUnexpected    Token = "<unexpected>"
	TokReturn        Token = "return"
	TokSingleComment Token = "#"
	TokMultiComment  Token = "%"
	TokNot           Token = "!"
	TokTrue          Token = "true"
	TokFalse         Token = "false"
	TokLShift        Token = "<<"
	TokRShift        Token = ">>"
	TokGt            Token = ">"
	TokGte           Token = ">="
	TokLt            Token = "<"
	TokLte           Token = "<="
	TokAnd           Token = "&&"
	TokBAnd          Token = "&"
	TokOr            Token = "||"
	TokBOr           Token = "|"
	TokXor           Token = "^"
	TokNEqual        Token = "!="
	TokIf            Token = "if"
	TokElse          Token = "else"
	TokFor           Token = "for"
	TokContinue      Token = "continue"
	TokBreak         Token = "break"
	TokExtern        Token = "extern"
	TokPublic        Token = "public"
	TokPrivate       Token = "private"
	TokStruct        Token = "struct"
	TokImpl          Token = "impl"
	TokNil           Token = "nil"
	TokSizeof        Token = "sizeof" // sizeof is a special unary operator that can be used to calculate the size of type in a compile time.
	TokDot           Token = "."
	TokLBracket      Token = "["
	TokRBracket      Token = "]"
	TokChar          Token = "<char_literal>"
	TokPackage       Token = "package"
	TokNew           Token = "new"
)
