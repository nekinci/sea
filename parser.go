package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Parser struct {
	*Lexer
	module *Package
	errors []string
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
		start:    0,
		inputLen: len(file),
	}

	parser.Lexer = lexer
	return parser
}

func (p *Parser) parse() *Package {
	pckg := &Package{
		Name:  "Main",
		Stmts: []Stmt{},
	}
	p.module = pckg

	p.next()
	for p.curTok != EOF {
		pckg.Stmts = append(pckg.Stmts, p.parseStmt())
	}

	return pckg

}

func (p *Parser) parseVar() *VarDefStmt {
	p.expect(TokVar)
	_, typ := p.expect(TokIdentifier)

	isPtr := false
	if p.curTok == TokMultiply {
		p.expect(TokMultiply)
		isPtr = true
	}

	_, name := p.expect(TokIdentifier)

	varDefStmt := &VarDefStmt{
		Name:  &IdentExpr{Name: name},
		Type:  &IdentExpr{Name: typ},
		Init:  nil,
		IsPtr: isPtr,
	}

	if p.curTok == TokAssign {
		p.expect(TokAssign)
		varDefStmt.Init = p.parseExpr()
	}
	return varDefStmt
}

func (p *Parser) parseIf() *IfStmt {
	p.expect(TokIf)
	ifStmt := &IfStmt{}
	ifStmt.Cond = p.parseExpr()
	ifStmt.Then = p.parseStmt()
	if p.curTok == TokElse {
		p.expect(TokElse)
		ifStmt.Else = p.parseStmt()
	}
	return ifStmt
}

func (p *Parser) parseField() *Field {
	_, typ := p.expect(TokIdentifier)
	_, name := p.expect(TokIdentifier)
	return &Field{
		Name: &IdentExpr{Name: name},
		Type: &IdentExpr{Name: typ},
	}
}

func (p *Parser) parseStruct() *StructDefStmt {
	p.expect(TokStruct)
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
	}
}

func (p *Parser) parseFor() *ForStmt {
	p.expect(TokFor)
	forStmt := &ForStmt{}

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
	case TokExtern:
		if p.mayNextBe(TokFun) {
			return p.parseFunc()
		} else if p.mayNextBe(TokVar) {
			return p.parseVar()
		}
		p.assume(false, "extern cannot be combined except fun and var")
	case TokFun:
		return p.parseFunc()
	case TokLBrace:
		return p.parseBlock()
	case TokBreak:
		p.expect(TokBreak)
		// TODO handle labels
		return &BreakStmt{}
	case TokContinue:
		p.expect(TokContinue)
		return &ContinueStmt{}
	case TokReturn:
		p.expect(TokReturn)
		returnStmt := &ReturnStmt{}
		returnStmt.Value = p.parseExpr()
		return returnStmt
	case TokVar:
		return p.parseVar()
	case TokIf:
		return p.parseIf()
	case TokFor:
		return p.parseFor()
	case TokStruct:
		return p.parseStruct()
	case TokImpl:
		return p.parseImpl()
	}

	return p.parseExprStmt()

}

func (p *Parser) parseImpl() *ImplStmt {
	panic("TODO")
}

func (p *Parser) parseBlock() *BlockStmt {
	p.expect(TokLBrace)
	blockStmt := &BlockStmt{Stmts: make([]Stmt, 0)}
	for {
		stmt := p.parseStmt()
		blockStmt.Stmts = append(blockStmt.Stmts, stmt)
		if p.curTok == TokRBrace {
			p.expect(TokRBrace)
			break
		}
	}

	return blockStmt
}

func (p *Parser) parseExprList() []Expr {
	exprs := make([]Expr, 0)

	for {
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
		left = &UnaryExpr{
			Op:    p.op(tok),
			Right: p.parseBinaryExpr(unaryPrecedence),
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

func (p *Parser) parseSelector() Expr {
	_, name := p.expect(TokIdentifier)
	var expr Expr
	expr = &IdentExpr{Name: name}
	// a.b.c
	//     ^
	// {Left: a, Right: {Left: b, Right: c}}
	// TODO think parens or something like
	for p.curTok == TokDot {
		p.expect(TokDot)
		r := p.parseSelector()
		expr = &SelectorExpr{Left: expr, Right: r}
	}

	return expr
}

func (p *Parser) parseSimpleExpr() Expr {
	switch p.curTok {
	case TokNil:
		p.expect(TokNil)
		return &NilExpr{}
	case TokString:
		_, str := p.expect(TokString)
		unquotedStr, err := strconv.Unquote(str)
		Assert(err == nil, "unquoted string expected")
		// implicitly add end of string to unquoted string
		unquotedStr += "\x00"
		return &StringExpr{Value: str, Unquoted: unquotedStr}
	case TokSizeof:
		p.expect(TokSizeof)
		p.expect(TokLParen)
		_, typ := p.expect(TokIdentifier)
		p.expect(TokLParen)
		// TODO Is UnaryExpr or special handling
		return &UnaryExpr{Op: Sizeof, Right: &IdentExpr{Name: typ}}
	case TokNumber:
		_, val := p.expect(TokNumber)
		v, err := strconv.ParseInt(val, 10, 64)
		Assert(err == nil, fmt.Sprintf("Invalid number: %v", err))
		return &NumberExpr{Value: v}
	case TokTrue, TokFalse:
		tok, _ := p.expectAnyOf(TokTrue, TokFalse)
		return &BoolExpr{Value: tok == TokTrue}
	case TokLParen:
		p.expect(TokLParen)
		expr := p.parseExpr()
		p.expect(TokRParen)
		return &ParenExpr{Expr: expr}
	case TokIdentifier:
		return p.parseIdentExpr()
	case TokMultiply:
		p.expect(TokMultiply)
		expr := p.parseSelector()
		expr = &UnaryExpr{Op: Mul, Right: expr}
		var right Expr
		if p.curTok == TokAssign {
			p.expect(TokAssign)
			right = p.parseExpr()
			return &AssignExpr{Left: expr, Right: right}
		}
		return expr
	case TokBAnd:
		p.expect(TokBAnd)
		expr := p.parseExpr()
		return &UnaryExpr{Op: Band, Right: expr}
	default:
		panic("unreachable tok -> " + string(p.curTok))

	}
}

func (p *Parser) parseIdentExpr() Expr {
	_, identifier := p.expect(TokIdentifier)
	if p.curTok == TokLParen {
		callExpr := &CallExpr{}
		callExpr.Name = &IdentExpr{Name: identifier}
		p.expect(TokLParen)
		callExpr.Args = p.parseExprList()
		p.expect(TokRParen)

		return callExpr
	}

	if p.curTok == TokAssign {
		p.expect(TokAssign)
		assignExpr := &AssignExpr{&IdentExpr{Name: identifier}, p.parseExpr()}
		return assignExpr
	}

	return &IdentExpr{Name: identifier}
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
	for p.curTok != TokRParen {
		_, typeVal := p.expect(TokIdentifier)
		_, identVal := p.expect(TokIdentifier)
		param := &ParamExpr{
			Name: &IdentExpr{Name: identVal},
			Type: &IdentExpr{Name: typeVal},
		}

		params = append(params, param)

		if p.curTok == TokComma {
			p.expect(TokComma)
		}

	}

	p.expect(TokRParen)
	return params
}

func (p *Parser) assume(cond bool, msg string) {
	if !cond {
		p.errors = append(p.errors, msg)
	}
}

func (p *Parser) expect(tok Token) (Token, string) {
	defer p.next()
	if p.curTok == tok {
		return p.curTok, p.curVal
	}

	p.assume(tok == p.curTok, fmt.Sprintf("Expected %s, got %s", tok, p.curTok))
	return TokUnexpected, p.curVal
}

func (p *Parser) expectAnyOf(tokens ...Token) (Token, string) {
	defer p.next()
	for _, tok := range tokens {
		if p.curTok == tok {
			return p.curTok, p.curVal
		}
	}

	p.assume(false, fmt.Sprintf("Expected %s, got %s", tokens, p.curTok))
	return TokUnexpected, p.curVal
}

func (p *Parser) mayNextBe(tok Token) bool {
	t, v := p.curTok, p.curVal
	b, _ := p.nextAndBackup()
	p.curTok, p.curVal = t, v
	return b == tok
}

func (p *Parser) parseFunc() *FuncDefStmt {

	var isExternal bool
	if p.curTok == TokExtern {
		isExternal = true
		p.expect(TokExtern)
	}

	p.expect(TokFun)

	funcDef := &FuncDefStmt{
		Params:     make([]*ParamExpr, 0),
		IsExternal: isExternal,
	}
	_, identifier := p.expect(TokIdentifier)
	funcDef.Type = &IdentExpr{Name: identifier}
	_, identifier = p.expect(TokIdentifier)
	funcDef.Name = &IdentExpr{Name: identifier}
	funcDef.Params = p.parseParams()
	if !isExternal {
		funcDef.Body = p.parseBlock()
	}
	return funcDef
}
