package main

type Package struct {
	Name  string
	Stmts []Stmt
}

type Expr interface {
	Node
	IsExpr()
}

type NumberExpr struct {
	Value int64
}

type ParenExpr struct {
	Expr Expr
}

func (p *ParenExpr) IsExpr() {}

type StringExpr struct {
	Value    string
	Unquoted string
}

func (s *StringExpr) IsExpr() {}

func (e *NumberExpr) IsExpr() {}

type BoolExpr struct {
	Value bool // 1 true 0 false
}

func (b *BoolExpr) IsExpr() {}

type UnaryExpr struct {
	Op    Operation
	Right Expr
}

type AssignExpr struct {
	Left  Expr
	Right Expr
}

func (a *AssignExpr) IsExpr() {}

func (u *UnaryExpr) IsExpr() {}

type BinaryExpr struct {
	Left  Expr
	Right Expr
	Op    Operation
}

func (e *BinaryExpr) IsExpr() {}

type CallExpr struct {
	Name *IdentExpr
	Args []Expr
}

func (c *CallExpr) IsExpr() {}

func isWhitespace(c uint8) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isDigit(c uint8) bool {
	return '0' <= c && c <= '9'
}

type IdentExpr struct {
	Name string
}

type SelectorExpr struct {
	Left  Expr
	Right Expr
}

func (e *SelectorExpr) IsExpr() {}

type NilExpr struct{}

func (e *NilExpr) IsNil()  {}
func (e *NilExpr) IsExpr() {}

func (e *IdentExpr) IsExpr() {}

type FuncDefStmt struct {
	Name       *IdentExpr
	Type       Expr
	Params     []*ParamExpr
	Body       *BlockStmt
	IsExternal bool
}

func (f *FuncDefStmt) IsStmt() {}
func (f *FuncDefStmt) IsDef()  {}

type Node interface{}

type ParamExpr struct {
	Name *IdentExpr
	Type Expr
}

type BlockStmt struct {
	Stmts []Stmt
}

type ImplStmt struct {
	Stmts   []Stmt
	Implies []*IdentExpr
}

func (e *ImplStmt) IsStmt() {}

type DefStmt interface {
	IsDef()
}

func (s *BlockStmt) IsStmt() {}

func (e *ParamExpr) IsExpr() {}

type Stmt interface {
	Node
	IsStmt()
}

type IfStmt struct {
	Cond Expr
	Then Stmt // what about single line un-blocked compileStmt?
	Else Stmt
}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Step Expr
	Body Stmt
}

func (f *ForStmt) IsStmt() {}

func (e *IfStmt) IsStmt() {}

type ReturnStmt struct {
	Value Expr
}

// BreakStmt TODO support labels
type BreakStmt struct {
}

func (b *BreakStmt) IsStmt() {}

type ContinueStmt struct{}

func (c *ContinueStmt) IsStmt() {}

func (e *ReturnStmt) IsStmt() {}

type VarDefStmt struct {
	Name  *IdentExpr
	Type  Expr
	IsPtr bool
	Init  Expr
}

func (e *VarDefStmt) IsDef() {}

func (e *VarDefStmt) IsStmt() {}

type StructDefStmt struct {
	Name   *IdentExpr
	Fields []*Field
}

func (s *StructDefStmt) IsDef() {}

func (s *StructDefStmt) IsStmt() {}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) IsStmt() {}

type Field struct {
	Name *IdentExpr
	Type *IdentExpr
}

func (f *Field) IsExpr() {}
