package main

type Pos struct {
	Col    int
	Line   int
	Offset int
}

type Node interface {
	Pos() (start Pos, end Pos)
}

type Package struct {
	Name  string
	Stmts []Stmt
}

func (p *Package) Pos() (Pos, Pos) {
	start, _ := p.Stmts[0].Pos()
	end, _ := p.Stmts[len(p.Stmts)-1].Pos()
	return start, end
}

type Expr interface {
	Node
	IsExpr()
}

type NumberExpr struct {
	Value      int64
	start, end Pos
}

func (n *NumberExpr) Pos() (start Pos, end Pos) {
	start = n.start
	end = n.end
	return start, end
}

type ParenExpr struct {
	Expr       Expr
	start, end Pos
}

func (p *ParenExpr) Pos() (Pos, Pos) {
	return p.start, p.end
}

func (p *ParenExpr) IsExpr() {}

type StringExpr struct {
	Value      string
	Unquoted   string
	start, end Pos
}

func (s *StringExpr) Pos() (Pos, Pos) {
	return s.start, s.end
}

func (s *StringExpr) IsExpr() {}

func (n *NumberExpr) IsExpr() {}

type BoolExpr struct {
	Value      bool // 1 true 0 false
	start, end Pos
}

func (b *BoolExpr) Pos() (Pos, Pos) { return b.start, b.end }

func (b *BoolExpr) IsExpr() {}

type UnaryExpr struct {
	Op    Operation
	Right Expr
	start Pos
}

func (u *UnaryExpr) Pos() (Pos, Pos) {
	_, end := u.Right.Pos()
	return u.start, end
}

type AssignExpr struct {
	Left  Expr
	Right Expr
}

func (a *AssignExpr) Pos() (Pos, Pos) {
	start, _ := a.Left.Pos()
	_, end := a.Right.Pos()
	return start, end
}

func (a *AssignExpr) IsExpr() {}

func (u *UnaryExpr) IsExpr() {}

type BinaryExpr struct {
	Left  Expr
	Right Expr
	Op    Operation
}

func (e *BinaryExpr) Pos() (Pos, Pos) {
	start, _ := e.Left.Pos()
	_, end := e.Right.Pos()
	return start, end
}

func (e *BinaryExpr) IsExpr() {}

type CallExpr struct {
	Left Expr
	Args []Expr
	end  Pos
}

func (c *CallExpr) Pos() (Pos, Pos) {
	start, _ := c.Left.Pos()
	end := c.end
	return start, end
}

func (c *CallExpr) IsExpr() {}

func isWhitespace(c uint8) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isDigit(c uint8) bool {
	return '0' <= c && c <= '9'
}

type IdentExpr struct {
	Name       string
	start, end Pos
}

// RefExpr used for type identifiers, like int*, bool* ...
type RefExpr struct {
	Expr Expr
	end  Pos
}

func (r *RefExpr) Pos() (Pos, Pos) {
	start, _ := r.Expr.Pos()
	return start, r.end
}

func (r *RefExpr) IsExpr() {}

func (e *IdentExpr) Pos() (Pos, Pos) {
	return e.start, e.end
}

type SelectorExpr struct {
	Ident    Expr
	Selector Expr
}

func (e *SelectorExpr) Pos() (Pos, Pos) {
	start, _ := e.Ident.Pos()
	_, end := e.Selector.Pos()
	return start, end
}

func (e *SelectorExpr) IsExpr() {}

type NilExpr struct {
	start, end Pos
}

func (e *NilExpr) Pos() (Pos, Pos) {
	return e.start, e.end
}

func (e *NilExpr) IsNil()  {}
func (e *NilExpr) IsExpr() {}

func (e *IdentExpr) IsExpr() {}

type FuncDefStmt struct {
	Name       *IdentExpr
	Type       Expr
	Params     []*ParamExpr
	Body       *BlockStmt
	IsExternal bool
	start      Pos
	end        *Pos
}

func (f *FuncDefStmt) Pos() (Pos, Pos) {
	start := f.start
	var end Pos
	if f.Body == nil {
		Assert(f.end != nil, "at least one position must be filled")
		end = *f.end
	} else {
		_, end = f.Body.Pos()
	}

	return start, end
}

func (f *FuncDefStmt) IsStmt() {}
func (f *FuncDefStmt) IsDef()  {}

type ParamExpr struct {
	Name *IdentExpr
	Type Expr
}

func (p *ParamExpr) Pos() (Pos, Pos) {
	start, _ := p.Type.Pos()
	_, end := p.Name.Pos()

	return start, end
}

type BlockStmt struct {
	Stmts      []Stmt
	start, end Pos
}

func (s *BlockStmt) Pos() (Pos, Pos) {
	return s.start, s.end
}

type ImplStmt struct {
	Stmts      []Stmt
	Implies    []*IdentExpr
	start, end Pos
}

func (e *ImplStmt) Pos() (Pos, Pos) {
	return e.start, e.end
}

func (e *ImplStmt) IsStmt() {}

type DefStmt interface {
	IsDef()
}

func (s *BlockStmt) IsStmt() {}

func (p *ParamExpr) IsExpr() {}

type Stmt interface {
	Node
	IsStmt()
}

type IfStmt struct {
	Cond  Expr
	Then  Stmt // what about single line un-blocked compileStmt?
	Else  Stmt
	start Pos
}

func (e *IfStmt) Pos() (Pos, Pos) {
	var end Pos
	if e.Else != nil {
		_, end = e.Else.Pos()
	} else {
		_, end = e.Then.Pos()
	}
	return e.start, end
}

type ForStmt struct {
	Init  Stmt
	Cond  Expr
	Step  Expr
	Body  Stmt
	start Pos
}

func (e *ForStmt) Pos() (Pos, Pos) {
	_, end := e.Body.Pos()
	return e.start, end
}

func (e *ForStmt) IsStmt() {}

func (e *IfStmt) IsStmt() {}

type ReturnStmt struct {
	Value      Expr
	start, end Pos
}

func (r *ReturnStmt) Pos() (Pos, Pos) {
	var end = r.end
	if r.Value != nil {
		_, end = r.Value.Pos()
	}
	return r.start, end
}

func (r *ReturnStmt) IsStmt() {}

// BreakStmt TODO support labels
type BreakStmt struct {
	start, end Pos
}

func (b *BreakStmt) Pos() (Pos, Pos) {
	return b.start, b.end
}

func (b *BreakStmt) IsStmt() {}

type ContinueStmt struct {
	start, end Pos
}

func (c *ContinueStmt) Pos() (Pos, Pos) {
	return c.start, c.end
}

func (c *ContinueStmt) IsStmt() {}

type VarDefStmt struct {
	Name  *IdentExpr
	Type  Expr
	Init  Expr
	start Pos
}

func (v *VarDefStmt) Pos() (Pos, Pos) {
	var end Pos
	if v.Init != nil {
		_, end = v.Init.Pos()
	} else {
		_, end = v.Name.Pos()
	}
	return v.start, end
}

func (v *VarDefStmt) IsDef() {}

func (v *VarDefStmt) IsStmt() {}

type StructDefStmt struct {
	Name       *IdentExpr
	Fields     []*Field
	start, end Pos
}

func (s *StructDefStmt) Pos() (Pos, Pos) {
	return s.start, s.end
}

func (s *StructDefStmt) IsDef() {}

func (s *StructDefStmt) IsStmt() {}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) Pos() (Pos, Pos) {
	return e.Expr.Pos()
}

func (e *ExprStmt) IsStmt() {}

type Field struct {
	Name *IdentExpr
	Type Expr
}

func (f *Field) Pos() (Pos, Pos) {
	start, _ := f.Name.Pos()
	_, end := f.Type.Pos()
	return start, end

}
func (f *Field) IsExpr() {}
