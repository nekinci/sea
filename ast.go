package main

type Pos struct {
	Col    int
	Line   int
	Offset int
}

type Node interface {
	Pos() (start Pos, end Pos)
}

type Constant interface {
	IsConstant() bool
}

type Contexter interface {
	setCtx(ctx Ctx)
	GetCtx() Ctx
}

type Package struct {
	Name      string
	Path      string
	Files     []*File
	Stmts     []Stmt
	FileMap   map[Stmt]string
	ImportMap map[*UseStmt]string
}

func (p *Package) AllStatements() []Stmt {
	if p.Stmts == nil {
		p.Stmts = make([]Stmt, 0)
		for _, file := range p.Files {
			for _, stmt := range file.Stmts {
				p.FileMap[stmt] = file.Name
				p.Stmts = append(p.Stmts, stmt)
				switch stmt := stmt.(type) {
				case *UseStmt:
					p.ImportMap[stmt] = file.Name
				}
			}
		}
	}

	return p.Stmts
}

type File struct {
	Name          string
	Stmts         []Stmt
	PackageStmt   *PackageStmt
	SmallestOrder int
}

func (p *File) Pos() (Pos, Pos) {
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
	ctx        Ctx
}

func (n *NumberExpr) IsConstant() bool { return true }

type FloatExpr struct {
	Value      float64
	start, end Pos
	ctx        Ctx
}

func (f *FloatExpr) IsConstant() bool { return true }

func (f *FloatExpr) setCtx(ctx Ctx) {
	f.ctx = ctx
}

func (f *FloatExpr) GetCtx() Ctx {
	return f.ctx
}

func (f *FloatExpr) Pos() (Pos, Pos) {
	return f.start, f.end
}

func (f *FloatExpr) IsExpr() {}

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
	Raw        string
	start, end Pos
	ctx        Ctx
}

func (s *StringExpr) IsConstant() bool { return true }

func (s *StringExpr) setCtx(ctx Ctx) {
	s.ctx = ctx
}

func (s *StringExpr) GetCtx() Ctx {
	return s.ctx
}

func (s *StringExpr) Pos() (Pos, Pos) {
	return s.start, s.end
}

func (s *StringExpr) IsExpr() {}

type CharExpr struct {
	Value      string
	Unquoted   rune
	start, end Pos
}

func (c *CharExpr) IsConstant() bool { return true }

func (c *CharExpr) Pos() (Pos, Pos) {
	return c.start, c.end
}

func (c *CharExpr) IsExpr() {}

func (n *NumberExpr) IsExpr() {}

func (n *NumberExpr) setCtx(ctx Ctx) {
	n.ctx = ctx
}

func (n *NumberExpr) GetCtx() Ctx {
	return n.ctx
}

type BoolExpr struct {
	Value      bool // 1 true 0 false
	start, end Pos
}

func (b *BoolExpr) IsConstant() bool { return true }

func (b *BoolExpr) Pos() (Pos, Pos) { return b.start, b.end }

func (b *BoolExpr) IsExpr() {}

type UnaryExpr struct {
	Op    Operation
	Right Expr
	start Pos
}

func (u *UnaryExpr) IsConstant() bool {
	r, ok := u.Right.(Constant)

	if ok {
		return r.IsConstant()
	}

	return false
}

func (u *UnaryExpr) Pos() (Pos, Pos) {
	_, end := u.Right.Pos()
	return u.start, end
}

type AssignExpr struct {
	Left  Expr
	Right Expr
	Op    string
	ctx   *VarAssignCtx
}

func (a *AssignExpr) setCtx(ctx Ctx) {
	ct, ok := ctx.(*VarAssignCtx)
	Assert(ok, "Unexpected ctx type")
	a.ctx = ct
}

func (a *AssignExpr) GetCtx() Ctx {
	return a.ctx
}

func (a *AssignExpr) Context() *VarAssignCtx {
	return a.ctx
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
	Ctx   *BinaryExprCtx
}

func (b *BinaryExpr) IsConstant() bool {

	l, ok1 := b.Left.(Constant)
	r, ok2 := b.Right.(Constant)
	if ok1 && ok2 {
		return l.IsConstant() && r.IsConstant()
	}

	return false
}

func (e *BinaryExpr) Pos() (Pos, Pos) {
	start, _ := e.Left.Pos()
	_, end := e.Right.Pos()
	return start, end
}

func (e *BinaryExpr) IsExpr() {}

type CallExpr struct {
	Left     Expr
	Args     []Expr
	end      Pos
	MethodOf string
	Package  string
	TypeCast bool
	ctx      Ctx
}

func (c *CallExpr) setCtx(ctx Ctx) {
	c.ctx = ctx
}

func (c *CallExpr) GetCtx() Ctx {
	return c.ctx
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

type ComplexLiteral interface {
	IsComplexLiteral()
}

type ArrayLitExpr struct {
	start, end Pos
	Elems      []Expr
	ctx        Ctx
}

func (a *ArrayLitExpr) Pos() (Pos, Pos) {
	return a.start, a.end
}
func (a *ArrayLitExpr) IsExpr()           {}
func (a *ArrayLitExpr) IsComplexLiteral() {}
func (a *ArrayLitExpr) setCtx(ctx Ctx) {
	a.ctx = ctx
}
func (a *ArrayLitExpr) GetCtx() Ctx {
	return a.ctx
}

// RefTypeExpr used for type identifiers, like int*, bool* ...
type RefTypeExpr struct {
	Expr Expr
	end  Pos
}

func (r *RefTypeExpr) Pos() (Pos, Pos) {
	start, _ := r.Expr.Pos()
	return start, r.end
}

func (r *RefTypeExpr) IsExpr() {}

func (e *IdentExpr) Pos() (Pos, Pos) {
	return e.start, e.end
}

type SelectorExpr struct {
	Ident    Expr
	Selector Expr
	Ctx      *SelectorCtx
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

type ArrayTypeExpr struct {
	Type Expr
	Size Expr
	end  Pos // RBracket pos
}

func (e *ArrayTypeExpr) Pos() (Pos, Pos) {
	start, _ := e.Type.Pos()
	return start, e.end
}

func (e *ArrayTypeExpr) IsExpr() {}

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
	ImplOf     *ImplStmt
	IsExternal bool
	start      Pos
	end        *Pos
	ctx        *FuncCtx
}

func (f *FuncDefStmt) setCtx(ctx Ctx) {
	ct, ok := ctx.(*FuncCtx)
	Assert(ok, "FuncCtx must be provided!")
	f.ctx = ct
}

func (f *FuncDefStmt) GetCtx() Ctx {
	return f.ctx
}

func (f *FuncDefStmt) Context() *FuncCtx {
	return f.ctx
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

type UseStmt struct {
	useCtx *UseCtx
	Path   *StringExpr
	Alias  *IdentExpr
	start  Pos
}

func (u *UseStmt) IsStmt() {}
func (u *UseStmt) Pos() (Pos, Pos) {
	if u.Alias != nil {
		_, end := u.Alias.Pos()
		return u.start, end
	}

	_, end := u.Path.Pos()
	return u.start, end
}

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
	Type       Expr
	Implies    []*IdentExpr
	RefOfType  DefStmt
	start, end Pos
}

func (e *ImplStmt) Pos() (Pos, Pos) {
	return e.start, e.end
}

func (e *ImplStmt) IsStmt() {}

type DefStmt interface {
	Stmt
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
	Step  Stmt
	Body  Stmt
	start Pos
	ctx   *ForCtx
}

func (e *ForStmt) Pos() (Pos, Pos) {
	_, end := e.Body.Pos()
	return e.start, end
}

func (e *ForStmt) IsStmt() {}

func (e *IfStmt) IsStmt() {}

type IncrStmt struct {
	Expr            Expr
	start, end      Pos
	IsPostOperation bool
	Type            string
}

type DecrStmt struct {
	Expr            Expr
	start, end      Pos
	IsPostOperation bool
	Type            string
}

func (i *DecrStmt) IsStmt() {}
func (i *DecrStmt) Pos() (Pos, Pos) {
	if i.IsPostOperation {
		start, _ := i.Expr.Pos()
		return start, i.end
	}
	_, end := i.Expr.Pos()
	return i.start, end
}

func (i *IncrStmt) IsStmt() {}
func (i *IncrStmt) Pos() (Pos, Pos) {
	if i.IsPostOperation {
		start, _ := i.Expr.Pos()
		return start, i.end
	}
	_, end := i.Expr.Pos()
	return i.start, end
}

type ReturnStmt struct {
	Value      Expr
	start, end Pos
	ctx        *FuncCtx
}

func getFuncCtx(ctx Ctx) *FuncCtx {
	if ct, ok := ctx.(*FuncCtx); ok {
		return ct
	}

	if ctx != nil && ctx.parentCtx() != nil {
		return getFuncCtx(ctx.parentCtx())
	}

	panic("invalid ctx")
}

func (r *ReturnStmt) setCtx(ctx Ctx) {
	funcCtx := getFuncCtx(ctx)
	r.ctx = funcCtx
}

func (r *ReturnStmt) GetCtx() Ctx {
	return r.ctx
}

func (r *ReturnStmt) Context() *FuncCtx {
	return r.ctx
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
	ctx        *ForCtx
}

func (b *BreakStmt) Pos() (Pos, Pos) {
	return b.start, b.end
}

func (b *BreakStmt) IsStmt() {}

type ContinueStmt struct {
	start, end Pos
	ctx        *ForCtx
}

func (c *ContinueStmt) Pos() (Pos, Pos) {
	return c.start, c.end
}

func (c *ContinueStmt) IsStmt() {}

type VarDefStmt struct {
	Name    *IdentExpr
	Type    Expr
	Init    Expr
	start   Pos
	Order   int
	IsConst bool
	ctx     *VarAssignCtx
}

func (v *VarDefStmt) setCtx(ctx Ctx) {
	ct, ok := ctx.(*VarAssignCtx)
	Assert(ok, "Unexpected ctx type")
	v.ctx = ct
}

func (v *VarDefStmt) GetCtx() Ctx {
	return v.ctx
}

func (v *VarDefStmt) Context() *VarAssignCtx {
	return v.GetCtx().(*VarAssignCtx)
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

type PackageStmt struct {
	Name  *IdentExpr
	start Pos
}

func (p *PackageStmt) IsStmt()         {}
func (p *PackageStmt) Pos() (Pos, Pos) { return p.start, p.Name.end }

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

type IndexExpr struct {
	Left       Expr
	Index      Expr
	start, end Pos
	ctx        *IndexCtx
}

func (i *IndexExpr) Pos() (Pos, Pos) {
	return i.start, i.end
}

func (i *IndexExpr) IsExpr() {}

type ObjectLitExpr struct {
	Type       Expr
	KeyValue   []*KeyValueExpr
	start, end Pos
}

func (o *ObjectLitExpr) Pos() (Pos, Pos) {
	return o.start, o.end
}

func (o *ObjectLitExpr) IsExpr()           {}
func (o *ObjectLitExpr) IsComplexLiteral() {}

type KeyValueExpr struct {
	Key   Expr
	Value Expr
	ctx   Ctx
}

func (kv *KeyValueExpr) setCtx(ctx Ctx) {
	kv.ctx = ctx
}

func (kv *KeyValueExpr) GetCtx() Ctx {
	return kv.ctx
}

func (kv *KeyValueExpr) Context() *VarAssignCtx {
	ctx, ok := kv.ctx.(*VarAssignCtx)
	Assert(ok, "expecting VarAssignCtx")
	return ctx
}

func (kv *KeyValueExpr) IsExpr() {}

func (kv *KeyValueExpr) Pos() (Pos, Pos) {
	start, _ := kv.Key.Pos()
	_, end := kv.Value.Pos()
	return start, end
}

type TryCatchStmt struct {
	TryBlock   Stmt
	CatchBlock *CatchClause
	start      Pos
}

type CatchClause struct {
	Block  *BlockStmt
	Params []*ParamExpr
	start  Pos
}

func (c *CatchClause) IsStmt() {}

func (c *CatchClause) Pos() (Pos, Pos) {
	_, end := c.Block.Pos()
	return c.start, end
}

func (t *TryCatchStmt) IsStmt() {}
func (t *TryCatchStmt) Pos() (Pos, Pos) {
	_, end := t.TryBlock.Pos()
	if t.CatchBlock != nil {
		_, end = t.CatchBlock.Block.Pos()
	}
	return t.start, end
}

type ThrowStmt struct {
	Arg   Expr
	start Pos
}

func (t *ThrowStmt) IsStmt() {}
func (t *ThrowStmt) Pos() (Pos, Pos) {
	_, end := t.Arg.Pos()
	return t.start, end
}
