package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type bracketMode uint8

const (
	bracketNone  bracketMode = iota
	bracketType  bracketMode = 1
	bracketIndex bracketMode = 2
)

type Error struct {
	Start, End Pos
	Message    string
	File       string
}

func (e Error) String() string {
	return fmt.Sprintf("%s:%d:%d: %s\n", e.File, e.Start.Line+1, e.Start.Col+1, e.Message)
}

type Parser struct {
	*Lexer
	module     *Package
	errors     []Error
	start, end Pos // indicates the start and end positions of the last expected token
}

func (p *Parser) newError(message string) Error {
	e := Error{
		Start:   p.Start(),
		Message: message,
		File:    p.filename,
	}
	e.End = p.End()
	return e
}

func (p *Parser) startOfLastExpected() Pos {
	return p.start
}

func (p *Parser) endOfLastExpected() Pos {
	return p.end
}

func NewParser(path string) *Parser {
	parser := &Parser{}
	Assert(path != "", "path cannot be empty")
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	lexer := &Lexer{
		filename: path,
		input:    string(file),
		pos:      0,
		inputLen: len(file),
	}

	parser.Lexer = lexer
	return parser
}

func (p *Parser) parse() (*Package, []Error) {
	pckg := &Package{
		Name:  "Main",
		Stmts: []Stmt{},
	}
	p.module = pckg

	p.next()
	for p.curTok != EOF {
		pckg.Stmts = append(pckg.Stmts, p.parseTopStmt())
	}

	return pckg, p.errors

}

func (p *Parser) parseTypeIdentExpr() Expr {
	expr := p.parseSelectorExpr(bracketType)
	if p.curTok == TokMultiply {
		p.expect(TokMultiply)
		return &RefTypeExpr{
			Expr: expr,
			end:  p.endOfLastExpected(),
		}
	}

	return expr
}

func (p *Parser) parseVar() *VarDefStmt {
	p.expect(TokVar)
	var start = p.startOfLastExpected()
	typExpr := p.parseTypeIdentExpr()

	nameExpr := p.parseIdentExpr()

	varDefStmt := &VarDefStmt{
		Name:  nameExpr,
		Type:  typExpr,
		Init:  nil,
		start: start,
	}

	if p.curTok == TokAssign {
		p.expect(TokAssign)
		expr := p.parseExpr()
		varDefStmt.Init = expr
	}
	return varDefStmt
}

func (p *Parser) parseIf() *IfStmt {
	p.expect(TokIf)
	ifStmt := &IfStmt{start: p.startOfLastExpected()}
	ifStmt.Cond = p.parseExpr()
	ifStmt.Then = p.parseStmt()
	if p.curTok == TokElse {
		p.expect(TokElse)
		ifStmt.Else = p.parseStmt()
	}
	return ifStmt
}

func (p *Parser) parseField() *Field {
	var typExpr = p.parseTypeIdentExpr()
	var nameExpr = p.parseIdentExpr()
	return &Field{
		Name: nameExpr,
		Type: typExpr,
	}
}

func (p *Parser) parseStruct() *StructDefStmt {
	p.expect(TokStruct)
	var start = p.startOfLastExpected()
	_, name := p.expect(TokIdentifier)
	p.expect(TokLBrace)
	fields := make([]*Field, 0)

	for p.curTok != TokRBrace {
		fields = append(fields, p.parseField())
	}

	p.expect(TokRBrace)

	return &StructDefStmt{
		Name:   &IdentExpr{Name: name},
		Fields: fields,
		start:  start,
		end:    p.endOfLastExpected(),
	}
}
func (p *Parser) parseFor() *ForStmt {
	p.expect(TokFor)
	forStmt := &ForStmt{start: p.startOfLastExpected()}

	// TODO add support multiple variable stmts
	if p.curTok == TokVar {
		varDefStmt := p.parseVar()
		forStmt.Init = varDefStmt
		p.expect(TokSemicolon)
	}

	forStmt.Cond = p.parseExpr()
	if p.curTok == TokSemicolon {
		p.expect(TokSemicolon)
		forStmt.Step = p.parseExpr()
	}

	forStmt.Body = p.parseStmt()

	return forStmt
}

func (p *Parser) parseStmt() Stmt {
	switch p.curTok {
	case TokLBrace:
		return p.parseBlock()
	case TokBreak:
		p.expect(TokBreak)
		// TODO handle labels
		return &BreakStmt{p.startOfLastExpected(), p.endOfLastExpected()}
	case TokContinue:
		p.expect(TokContinue)
		return &ContinueStmt{p.startOfLastExpected(), p.endOfLastExpected()}
	case TokReturn:
		p.expect(TokReturn)
		returnStmt := &ReturnStmt{start: p.startOfLastExpected()}
		if p.Start().Line == returnStmt.start.Line {
			returnStmt.Value = p.parseExpr()
		} else {
			returnStmt.end = p.endOfLastExpected()
		}
		return returnStmt
	case TokIf:
		return p.parseIf()
	case TokFor:
		return p.parseFor()
	case TokVar:
		return p.parseVar()
	default:
		return p.parseExprStmt()
	}
}

func (p *Parser) parseTopStmt() Stmt {

	switch p.curTok {
	case TokExtern:
		if p.mayNextBe(TokFun) {
			return p.parseFunc()
		} else if p.mayNextBe(TokVar) {
			return p.parseVar()
		}
		p.expectAnyOf(TokFun, TokVar)
		return nil
	case TokFun:
		return p.parseFunc()
	case TokVar:
		return p.parseVar()
	case TokStruct:
		return p.parseStruct()
	case TokImpl:
		return p.parseImpl()
	default:
		p.expectAnyOf(TokFun, TokExtern, TokVar, TokStruct, TokImpl)
		return nil
	}

}

func (p *Parser) parseImpl() *ImplStmt {
	p.expect(TokImpl)
	implStmt := &ImplStmt{start: p.startOfLastExpected()}
	implStmt.Type = p.parseTypeIdentExpr()
	p.expect(TokLBrace)
	implStmt.Stmts = make([]Stmt, 0)

	if p.curTok == TokColon {
		panic("Unhandled yet tok colon on impl block")
	}

	for p.curTok != TokRBrace {
		stmt := p.parseFunc()
		implStmt.Stmts = append(implStmt.Stmts, stmt)
	}
	p.expect(TokRBrace)
	implStmt.end = p.endOfLastExpected()
	return implStmt
}

func (p *Parser) parseBlock() *BlockStmt {
	p.expect(TokLBrace)
	blockStmt := &BlockStmt{Stmts: make([]Stmt, 0), start: p.startOfLastExpected()}
	for p.curTok != EOF {
		stmt := p.parseStmt()
		blockStmt.Stmts = append(blockStmt.Stmts, stmt)
		if p.curTok == TokRBrace {
			p.expect(TokRBrace)
			blockStmt.end = p.endOfLastExpected()
			break
		}
	}
	return blockStmt
}

func (p *Parser) parseExprList() []Expr {
	exprs := make([]Expr, 0)

	for p.curTok != TokRParen {
		expr := p.parseExpr()
		exprs = append(exprs, expr)
		if p.curTok == TokComma {
			p.expect(TokComma)
		} else {
			break
		}
	}

	return exprs
}

func (p *Parser) binaryPrecedence(tok Token) int {

	// If it is binary expression then minimum binaryPrecedence is 1; otherwise return 0 that indicates it is not binary operator
	// Operation precedence
	// 1. Multiply, Division, Lshift, Rshift, Mod
	// 2. Plus, Minus
	// 3. gt, lt, gte, lte,neq, eq
	// 4. and, band
	// 5. or, bor, xor
	// And also unary operators are more precise

	// TODO paren case is important, in paren case that is should be skipped
	if p.line > p.lastLine /* && p.isInsideParen() */ {
		return 0
	}

	switch tok {
	case TokMultiply, TokDivision, TokLShift, TokRShift, TokMod:
		return 5
	case TokPlus, TokMinus:
		return 4
	case TokGt, TokGte, TokLt, TokLte, TokNEqual, TokEqual:
		return 3
	case TokAnd, TokBAnd:
		return 2
	case TokOr, TokBOr, TokXor:
		return 1
	}

	return 0
}

func (p *Parser) unaryPrecedence(tok Token) int {
	switch tok {
	case TokMinus, TokPlus, TokNot:
		return 6
	default:
		return 0
	}
}

func (p *Parser) parseBinaryExpr(parentPrecedence int) Expr {
	// 5 + 4 * 3
	// BinaryExpr[5 + BinaryExpr[4*3]]

	// 5 * 4 + 3
	// BinaryExpr[BinaryExpr[5*4] + 3]

	// 5*4*3
	// BinaryExpr[BinaryExpr[5*4] * 3]

	// -5+3
	var left Expr

	unaryPrecedence := p.unaryPrecedence(p.curTok)
	if unaryPrecedence != 0 && unaryPrecedence >= parentPrecedence {
		tok, _ := p.expectAnyOf(TokMinus, TokNot, TokPlus)
		var start = p.startOfLastExpected()
		left = &UnaryExpr{
			Op:    p.op(tok),
			Right: p.parseBinaryExpr(unaryPrecedence),
			start: start,
		}
	} else {
		left = p.parseSimpleExpr()
	}

	precedence := p.binaryPrecedence(p.curTok)

	for precedence > 0 && precedence > parentPrecedence {
		var op = p.op(p.curTok)
		p.next()
		left = &BinaryExpr{
			Left:  left,
			Right: p.parseBinaryExpr(precedence),
			Op:    op,
		}
		precedence = p.binaryPrecedence(p.curTok)

	}

	return left
}

func (p *Parser) op(tok Token) Operation {
	switch tok {
	case TokPlus:
		return Add
	case TokMinus:
		return Sub
	case TokDivision:
		return Div
	case TokMultiply:
		return Mul
	case TokEqual:
		return Eq
	case TokGt:
		return Gt
	case TokGte:
		return Gte
	case TokLt:
		return Lt
	case TokLte:
		return Lte
	case TokNEqual:
		return Neq
	case TokAnd:
		return And
	case TokOr:
		return Or
	case TokBAnd:
		return Band
	default:
		panic("unreachable op=" + string(tok))
	}
}

func (p *Parser) parseIdentExpr() *IdentExpr {
	_, name := p.expect(TokIdentifier)
	expr := &IdentExpr{Name: name, start: p.startOfLastExpected(), end: p.endOfLastExpected()}
	return expr
}

func (p *Parser) parseSelectorExpr(lbracketMode bracketMode) Expr {

	var expr Expr = p.parseIdentExpr()
	// a.b.c.d
	//       ^
	// %1 = {Ident: b, Selector: a}
	// %2 = {Ident: c, Selector: %1}
	// %3 = {Ident: d, Selector: %2}
	// {Ident: d, Selector: {Ident: c, Selector: {Ident: b, Selector: a}}}

	// a.b.c().d
	// %1 = {Ident: b, Selector: a}
	// %2 = {Ident: c, Selector: %1}
	// %3 = {Left: %2, Args:...}
	// %4 = {Ident: d, Selector: %3}
	// {Ident: d, Selector: {Left: {Ident: c, Selector: {Ident: b, Selector: A}}, Args:...}}

	// b()
	// %1 = {Left: b, Args:...}

	// b().c().d
	// %1 = {Left: b, Args:...}
	// %2 = {Ident: c, Selector: %1}
	// %3 = {Left: %2, Args:...}
	// %4 = {Ident: d, Selector: %3}
	// {Ident: d, Selector: {Left: {Left: {Ident: c, Selector: {Left: b, Args:...} }, Args:...} , Args:...}}

	for p.curTok == TokDot || p.curTok == TokLParen || p.curTok == TokLBracket {
		if p.curTok == TokLParen {
			p.expect(TokLParen)
			callExpr := &CallExpr{}
			callExpr.Left = expr
			callExpr.Args = p.parseExprList()
			p.expect(TokRParen)
			callExpr.end = p.endOfLastExpected()
			expr = callExpr
		} else if p.curTok == TokDot {
			p.expect(TokDot)
			expr = &SelectorExpr{Ident: p.parseIdentExpr(), Selector: expr}
		} else if p.curTok == TokLBracket {
			p.expect(TokLBracket)
			switch lbracketMode {
			case bracketType:
				arrayType := &ArrayTypeExpr{
					Type: expr,
					end:  Pos{},
				}
				if p.curTok != TokRBracket {
					arrayType.Size = p.parseNumberExpr()
				}
				p.expect(TokRBracket)
				arrayType.end = p.endOfLastExpected()
				return arrayType
			case bracketIndex:
				indexExpr := &IndexExpr{start: p.startOfLastExpected(), Left: expr}
				indexExpr.Index = p.parseExpr()
				p.expect(TokRBracket)
				indexExpr.end = p.endOfLastExpected()
				expr = indexExpr
			default:
				panic("unreachable lbracket mode")
			}
		}
	}

	// Is it fit in here?
	if p.curTok == TokLBrace {
		p.expect(TokLBrace)
		objLit := &ObjectLitExpr{
			Type:  expr,
			start: p.startOfLastExpected(),
		}
		keyValues := make([]*KeyValueExpr, 0)
		for p.curTok != TokRBrace {
			kv := p.parseKeyValueExpr()
			keyValues = append(keyValues, kv)
			if p.curTok != TokComma {
				break
			} else {
				p.expect(TokComma)
			}
		}
		p.expect(TokRBrace)
		objLit.KeyValue = keyValues
		objLit.end = p.endOfLastExpected()
		expr = objLit
	}

	return expr
}

func (p *Parser) parseNumberExpr() *NumberExpr {
	_, val := p.expect(TokNumber)
	v, err := strconv.ParseInt(val, 10, 64)
	Assert(err == nil, fmt.Sprintf("Invalid number: %v", err))
	return &NumberExpr{Value: v, start: p.startOfLastExpected(), end: p.endOfLastExpected()}
}

func (p *Parser) parseKeyValueExpr() *KeyValueExpr {
	kv := &KeyValueExpr{}
	kv.Key = p.parseIdentExpr()
	p.expect(TokColon)
	kv.Value = p.parseExpr()
	return kv
}

func (p *Parser) parseSimpleExpr() Expr {
	switch p.curTok {
	case TokLBracket:
		p.expect(TokLBracket)
		var elems = make([]Expr, 0)
		arrayValue := &ArrayLitExpr{start: p.startOfLastExpected()}
		for p.curTok != TokRBracket {
			elems = append(elems, p.parseExpr())
			if p.curTok == TokComma {
				p.expect(TokComma)
			}
		}
		arrayValue.Elems = elems
		p.expect(TokRBracket)
		arrayValue.end = p.endOfLastExpected()
		return arrayValue
	case TokNil:
		p.expect(TokNil)
		return &NilExpr{p.startOfLastExpected(), p.endOfLastExpected()}
	case TokString:
		_, str := p.expect(TokString)
		unquotedStr, err := strconv.Unquote(str)
		Assert(err == nil, "unquoted string expected")
		// implicitly add end of string to unquoted string
		unquotedStr += "\x00"
		return &StringExpr{Value: str, Unquoted: unquotedStr, start: p.startOfLastExpected(), end: p.endOfLastExpected()}
	case TokSizeof:
		p.expect(TokSizeof)
		var start = p.startOfLastExpected()
		p.expect(TokLParen)
		_, typ := p.expect(TokIdentifier)
		identStart, identEnd := p.startOfLastExpected(), p.endOfLastExpected()
		p.expect(TokRParen)
		// TODO Is UnaryExpr or special handling
		return &UnaryExpr{Op: Sizeof, Right: &IdentExpr{Name: typ, start: identStart, end: identEnd}, start: start}
	case TokNumber:
		return p.parseNumberExpr()
	case TokTrue, TokFalse:
		tok, _ := p.expectAnyOf(TokTrue, TokFalse)
		return &BoolExpr{Value: tok == TokTrue, start: p.startOfLastExpected(), end: p.endOfLastExpected()}
	case TokLParen:
		p.expect(TokLParen)
		start := p.startOfLastExpected()
		expr := p.parseExpr()
		p.expect(TokRParen)
		return &ParenExpr{Expr: expr, start: start, end: p.endOfLastExpected()}
	case TokIdentifier:
		var identExpr = p.parseSelectorExpr(bracketIndex)
		if p.curTok == TokAssign {
			return p.parseAssignExpr(identExpr)
		}
		return identExpr
	case TokMultiply:
		p.expect(TokMultiply)
		var start = p.startOfLastExpected()
		expr := p.parseSelectorExpr(bracketIndex)
		expr = &UnaryExpr{Op: Mul, Right: expr, start: start}
		var right Expr
		if p.curTok == TokAssign {
			p.expect(TokAssign)
			right = p.parseExpr()
			return &AssignExpr{Left: expr, Right: right}
		}
		return expr
	case TokBAnd:
		p.expect(TokBAnd)
		var start = p.startOfLastExpected()
		expr := p.parseExpr()
		return &UnaryExpr{Op: Band, Right: expr, start: start}
	default:
		p.errorf(false, "Unsupported token: "+p.curVal)
		p.next()
		return nil

	}
}

func (p *Parser) parseAssignExpr(identExpr Expr) *AssignExpr {
	p.expect(TokAssign)
	assignExpr := &AssignExpr{identExpr, p.parseExpr()}
	return assignExpr
}

func (p *Parser) parseExpr() Expr {

	switch p.curTok {
	default:
		return p.parseBinaryExpr(0)
	}

}

func (p *Parser) parseExprStmt() *ExprStmt {
	return &ExprStmt{Expr: p.parseExpr()}
}

func (p *Parser) parseParams() []*ParamExpr {
	p.expect(TokLParen)
	params := make([]*ParamExpr, 0)
	for p.curTok != TokRParen && p.curTok != EOF {
		typExpr := p.parseTypeIdentExpr()
		nameExpr := p.parseIdentExpr()
		param := &ParamExpr{
			Name: nameExpr,
			Type: typExpr,
		}

		params = append(params, param)

		if p.curTok == TokComma {
			p.expect(TokComma)
		}

	}

	p.expect(TokRParen)
	return params
}

func (p *Parser) errorf(cond bool, msg string) {
	if !cond {
		p.errors = append(p.errors, p.newError(msg))
	}
}

// expect checks the given token either matches the current token and then advances to the next token
// note that, Start() and End() methods should be called before the expect() method to ensure that get correct position
func (p *Parser) expect(tok Token) (Token, string) {
	defer p.next()
	p.start = p.Start()
	p.end = p.End()
	if p.curTok == tok {
		return p.curTok, p.curVal
	}

	p.errorf(tok == p.curTok, fmt.Sprintf("Expected %s, got %s", tok, p.curTok))
	return TokUnexpected, p.curVal
}

func (p *Parser) expectAnyOf(tokens ...Token) (Token, string) {
	defer p.next()
	for _, tok := range tokens {
		if p.curTok == tok {
			return p.curTok, p.curVal
		}
	}

	p.errorf(false, fmt.Sprintf("Expected %s, got %s", tokens, p.curTok))
	return TokUnexpected, p.curVal
}

func (p *Parser) mayNextBe(tok Token) bool {
	t, v := p.curTok, p.curVal
	b, _ := p.nextAndBackup()
	p.curTok, p.curVal = t, v
	return b == tok
}

func (p *Parser) parseFunc() *FuncDefStmt {
	var start = p.Start()
	var isExternal bool
	if p.curTok == TokExtern {
		isExternal = true
		p.expect(TokExtern)
	}

	p.expect(TokFun)

	funcDef := &FuncDefStmt{
		Params:     make([]*ParamExpr, 0),
		IsExternal: isExternal,
		start:      start,
	}
	_, identifier := p.expect(TokIdentifier)
	funcDef.Type = &IdentExpr{Name: identifier}
	tok, identifier := p.expect(TokIdentifier)
	if tok == TokUnexpected {
		return nil
	}
	funcDef.Name = &IdentExpr{Name: identifier}
	funcDef.Params = p.parseParams()
	if !isExternal {
		funcDef.Body = p.parseBlock()
	} else {
		var end = p.End()
		funcDef.end = &end
	}
	return funcDef
}
