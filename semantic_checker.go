package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const arrayScheme = "array<%s, %d>"
const pointerScheme = "pointer<%s>"
const unresolvedType = "<unresolved>"
const autoCastScheme = "auto_cast<%s>"

type Symbol interface {
	IsSymbol()
	Pos() (start Pos, end Pos)
	IsVarDef() bool
	IsTypeDef() bool
	IsFuncDef() bool
	TypeName() string
}

type TypeField struct {
	Node       *Field
	Name, Type string
}

type Param struct {
	Param *ParamExpr
	Name  string
	Type  string
}

type TypeDef struct {
	DefNode   DefStmt
	Name      string
	Fields    []TypeField
	Methods   []any
	Completed bool
	IsStruct  bool
}

func (t *TypeDef) TypeName() string {
	return t.Name
}

func (t *TypeDef) IsSymbol() {}
func (t *TypeDef) IsTypeDef() bool {
	return true
}
func (t *TypeDef) IsVarDef() bool {
	return false
}

func (t *TypeDef) IsFuncDef() bool {
	return false
}

func (t *TypeDef) Pos() (Pos, Pos) {
	if t.DefNode == nil {
		return Pos{}, Pos{} // for built-in types
	}
	return t.DefNode.Pos()
}

type VarDef struct {
	DefNode *VarDefStmt
	Name    string
	Type    string
}

func (v *VarDef) IsSymbol()       {}
func (v *VarDef) IsVarDef() bool  { return true }
func (v *VarDef) IsTypeDef() bool { return false }
func (v *VarDef) IsFuncDef() bool { return false }
func (v *VarDef) Pos() (Pos, Pos) { return v.DefNode.Pos() }
func (v *VarDef) TypeName() string {
	return v.Type
}

type FuncDef struct {
	DefNode   *FuncDefStmt
	Name      string
	Type      string // indicates return type
	Params    []Param
	MethodOf  string
	External  bool
	Variadic  bool
	Completed bool
	TypeCast  bool
}

func (f *FuncDef) TypeName() string {
	return f.Type
}
func (f *FuncDef) IsSymbol()       {}
func (f *FuncDef) IsVarDef() bool  { return false }
func (f *FuncDef) IsTypeDef() bool { return false }
func (f *FuncDef) IsFuncDef() bool { return true }
func (f *FuncDef) Pos() (Pos, Pos) {
	if f.DefNode == nil {
		return Pos{}, Pos{} // for built-in functions
	}
	return f.DefNode.Pos()

}

type Scope struct {
	Symbols    map[string]Symbol
	Parent     *Scope
	Children   []*Scope
	funcScopes map[string]*Scope
}

func (c *Checker) EnterScope() *Scope {
	newScope := &Scope{
		Symbols:    make(map[string]Symbol),
		funcScopes: make(map[string]*Scope),
	}

	if c.Scope != nil {
		c.Scope.Children = append(c.Scope.Children, newScope)
		newScope.Parent = c.Scope
	}

	c.Scope = newScope
	return newScope
}

func (c *Checker) enterTypeScope(name string) *Scope {
	if c.typeScopes[name] == nil {
		c.typeScopes[name] = c.EnterScope()
	}

	c.tmpScope = c.Scope
	c.Scope = c.typeScopes[name]
	return c.typeScopes[name]
}

func (c *Checker) enterFuncScope(name string) *Scope {
	if c.funcScopes[name] == nil {
		c.funcScopes[name] = c.EnterScope()
	}

	return c.funcScopes[name]
}

func (c *Checker) CloseScope() *Scope {
	c.Scope = c.Parent
	return c.Scope
}

type Checker struct {
	*Scope
	tmpScope    *Scope
	Package     *Package
	Errors      []Error
	currentFile string
	typeScopes  map[string]*Scope

	// Context
	currentSym Symbol
	ctx        Ctx
	//
}

func (c *Checker) newCtx(ctx Ctx) {
	c.ctx = ctx
}

func (c *Checker) leaveCtx() {
	c.ctx = c.ctx.parentCtx()
}

func (c *Checker) turnTempScopeBack() {
	c.Scope = c.tmpScope
}

func (c *Checker) Check() ([]Error, bool) {

	c.EnterScope()
	c.typeScopes = make(map[string]*Scope)
	defer c.CloseScope()
	c.initGlobalScope()
	c.collectSignatures()
	c.check()
	return c.Errors, len(c.Errors) > 0
}

func (c *Checker) addSymbol(name string, sym Symbol) error {
	if _, ok := c.Scope.Symbols[name]; ok {
		return fmt.Errorf("duplicate symbol %s", name)
	}
	c.Scope.Symbols[name] = sym
	return nil
}

func (c *Checker) addBuiltinTypes(names ...string) {
	for _, name := range names {
		_ = c.addSymbol(name, &TypeDef{
			DefNode:   nil,
			Name:      name,
			Fields:    nil,
			Methods:   nil,
			Completed: true,
		})

		if name != "void" {
			_ = c.addSymbol("cast_"+name, &FuncDef{
				DefNode:   nil,
				Name:      "cast_" + name,
				Type:      name,
				Params:    []Param{},
				MethodOf:  "",
				External:  false,
				Variadic:  false,
				Completed: true,
				TypeCast:  true,
			})
		}
	}
}

func (c *Checker) initGlobalScope() {

	c.addBuiltinTypes("string", "char", "bool", "i64", "i32", "i16", "i8", "f16", "f32", "f64", "void")

	err := c.addSymbol("printf_internal", &FuncDef{
		DefNode: nil,
		Name:    "printf_internal",
		Type:    "i32",
		Params: []Param{{
			Param: nil,
			Name:  "fmt",
			Type:  "pointer<char>",
		}},
		External:  true,
		Variadic:  true,
		Completed: false,
	})
	AssertErr(err)
	err = c.addSymbol("scanf_internal", &FuncDef{
		DefNode: nil,
		Name:    "scanf_internal",
		Type:    "i32",
		Params: []Param{{
			Param: nil,
			Name:  "fmt",
			Type:  "string",
		}},
		External:  true,
		Variadic:  true,
		Completed: false,
	})
	AssertErr(err)

	err = c.addSymbol("malloc_internal", &FuncDef{
		DefNode: nil,
		Name:    "malloc_internal",
		Type:    "pointer<i8>",
		Params: []Param{{
			Param: nil,
			Name:  "size",
			Type:  "i64",
		}},
		External:  true,
		Variadic:  false,
		Completed: true,
		TypeCast:  false,
	})
	AssertErr(err)
}

func (scope *Scope) LookupSym(name string) Symbol {
	if sym, ok := scope.Symbols[name]; ok {
		return sym
	}

	if scope.Parent != nil {
		return scope.Parent.LookupSym(name)
	}

	return nil
}

func (scope *Scope) GetSymbol(name string) Symbol {
	if sym, ok := scope.Symbols[name]; ok {
		return sym
	}

	return nil
}

func (c *Checker) DefineVar(name string, typ string) {
	if sym := c.GetSymbol(name); sym != nil {
		start, end := sym.Pos()
		c.errorf(start, end, "re defined variable %s", name)
	}

	_ = c.addSymbol(name, &VarDef{
		DefNode: nil,
		Name:    name,
		Type:    typ,
	})
}

func (c *Checker) getNameOfType(expr Expr) string {
	switch expr := expr.(type) {
	case *IdentExpr:
		return expr.Name
	case *RefTypeExpr:
		return fmt.Sprintf(pointerScheme, c.getNameOfType(expr.Expr))
	case *ArrayTypeExpr:
		var size = -1
		if expr.Size != nil {
			switch e := expr.Size.(type) {
			case *NumberExpr:
				size = int(e.Value)
			default:
				panic("Unhandled size")
			}
		}
		return fmt.Sprintf(arrayScheme, c.getNameOfType(expr.Type), size)
	default:
		panic("unreachable")
	}
}

func (c *Checker) errorf(start, end Pos, format string, args ...interface{}) {
	c.Errors = append(c.Errors, Error{
		Start:   start,
		End:     end,
		Message: fmt.Sprintf(format, args...),
		File:    "./input.sea",
	})
}

func (c *Checker) isPointer(name string) bool {
	return strings.HasPrefix(name, "pointer<")
}

func (c *Checker) checkStructDef(stmt *StructDefStmt) {
	c.enterTypeScope(stmt.Name.Name)
	defer c.CloseScope()
	sym := c.LookupSym(stmt.Name.Name)
	Assert(sym.IsTypeDef(), "Invalid definition of type")
	structDef := sym.(*TypeDef)
	structDef.IsStruct = true
	Assert(structDef.DefNode == stmt, "Invalid definition of type")
	defer func() {
		structDef.Completed = true
	}()

	if structDef.Completed {
		return
	}

	var nameUsed = make(map[string]uint)
	for _, field := range structDef.Fields {
		if count, ok := nameUsed[field.Name]; ok && count > 0 {
			start, end := field.Node.Name.Pos()
			c.errorf(start, end, "Cannot use same name %d times", count+1)
			nameUsed[field.Name] = nameUsed[field.Name] + 1
		} else {
			nameUsed[field.Name] = 1
			c.DefineVar(field.Name, field.Type)
		}

		typ := c.extractBaseType(field.Type)
		fieldType := c.LookupSym(typ)

		if fieldType == nil || !fieldType.IsTypeDef() {
			start, end := field.Node.Type.Pos()
			c.errorf(start, end, "type %s is not defined", field.Type)
		}

		if fieldType != nil && fieldType.TypeName() == stmt.Name.Name && !c.isPointer(field.Type) {
			start, end := stmt.Pos()
			c.errorf(start, end, "invalid recursive type %s. you may consider change field type like this: %s*", stmt.Name.Name, stmt.Name.Name)
		}

	}
}

func (c *Checker) GetVariable(name string) Symbol {
	if sym, ok := c.Scope.Symbols[name]; ok {
		return sym
	}
	return nil
}

// extractType extracts wrapped type only once
func (c *Checker) extractType(name string) string {
	if strings.HasPrefix(name, "auto_cast<") {
		var s string
		fmt.Sscanf(name, autoCastScheme, &s)
		s = strings.TrimSuffix(s, ">")
		return s
	} else if strings.HasPrefix(name, "pointer<") {
		var s string
		fmt.Sscanf(name, pointerScheme, &s)
		s = strings.TrimSuffix(s, ">")
		return s
	} else if strings.HasPrefix(name, "array<") {
		var s string
		fmt.Sscanf(name, arrayScheme, &s)
		s = strings.TrimSuffix(s, ">")
		s = strings.TrimSuffix(s, ",")
		return s
	}

	return name
}

func (c *Checker) getArraySizeFromTypeStr(typ string) int {
	if !strings.HasPrefix(typ, "array<") {
		return -1
	}
	s := strings.TrimPrefix(typ, "array<")
	s = strings.TrimSuffix(s, ">")
	x := strings.SplitAfter(s, ",")
	atoi, err := strconv.Atoi(strings.TrimSpace(x[1]))
	if err != nil {
		return -2
	}

	return atoi

}

func (c *Checker) isArray(name string) bool {
	return strings.HasPrefix(name, "array<")
}

func (c *Checker) checkTypeCompatibility(typ1, exprType string) bool {
	if strings.HasPrefix(exprType, "auto_cast<") {
		n := c.extractType(exprType)
		if c.isPointer(n) {
			return true
		}
		panic("not implemented auto casting case")
	}

	if exprType == "pointer<nil>" && c.isPointer(typ1) {
		return true
	}

	if c.isArray(exprType) && c.isArray(typ1) {
		typ1Size := c.getArraySizeFromTypeStr(typ1)
		exprTypeSize := c.getArraySizeFromTypeStr(exprType)

		typ1Base := c.extractBaseType(typ1)
		exprTypeBase := c.extractBaseType(exprType)
		if typ1Base != exprTypeBase {
			return false
		}

		if typ1Size == -1 || typ1Size == exprTypeSize {
			return true
		}

	}

	return typ1 == exprType
}

// getBitSize returns expectedBitSize and if it is not matched any bit-typed type returns -1.
func getBitSize(typName string) int {
	if typName == "i32" || typName == "f32" {
		return 32
	} else if typName == "i16" || typName == "f16" {
		return 16
	} else if typName == "i8" {
		return 8
	} else if typName == "f64" || typName == "i64" {
		return 64
	}

	return -1
}

// extractBaseType extracts type until base type
func (c *Checker) extractBaseType(name string) string {

	if strings.HasPrefix(name, "pointer<") {
		return c.extractBaseType(c.extractType(name))
	} else if strings.HasPrefix(name, "auto_cast<") {
		return c.extractType(c.extractType(name))
	} else if strings.HasPrefix(name, "array<") {
		return c.extractType(c.extractType(name))
	}

	// it is already unwrapped so
	return name
}

func (c *Checker) fixVarDefStmt(stmt *VarDefStmt, l, r string, ctx *VarAssignCtx) {
	if c.isArray(l) && c.isArray(r) {
		if c.getArraySizeFromTypeStr(l) == -1 {
			stmt.Type.(*ArrayTypeExpr).Size = &NumberExpr{Value: int64(c.getArraySizeFromTypeStr(r))}
		}

		stmt.Init.(*ArrayLitExpr).setCtx(ctx)
		ctx.arraySize = c.getArraySizeFromTypeStr(r)
	}
}

func (c *Checker) checkVarDef(stmt *VarDefStmt) {
	variable := c.GetVariable(stmt.Name.Name)
	if variable != nil {
		start, end := stmt.Name.Pos()
		c.errorf(start, end, "redeclared symbol %s", stmt.Name.Name)
	}

	typName := c.getNameOfType(stmt.Type)
	var typ = c.extractBaseType(typName)
	typSym := c.LookupSym(typ)
	ctx := &VarAssignCtx{
		parent:       c.ctx,
		expectedType: typName,
	}

	c.newCtx(ctx)
	defer c.leaveCtx()
	stmt.setCtx(ctx)

	if typSym == nil || !typSym.IsTypeDef() {
		start, end := stmt.Type.Pos()
		c.errorf(start, end, "type %s is not defined", typName)
	}

	if stmt.Init != nil {

		r, err := c.checkExpr(stmt.Init)
		if err != nil {
			return
		}

		if typSym != nil {
			if !c.checkTypeCompatibility(typName, r) {
				start, end := stmt.Init.Pos()
				c.errorf(start, end, "Expected type %s but got %s", typName, r)
			} else {
				c.fixVarDefStmt(stmt, typName, r, ctx)
				c.setVarAssignCtxFields(ctx, typName, typSym.(*TypeDef))
			}
		}

	}

	c.DefineVar(stmt.Name.Name, typName)
}

func (c *Checker) setVarAssignCtxFields(ctx *VarAssignCtx, typName string, typSym *TypeDef) {
	ctx.isStruct = typSym.IsStruct
	ctx.isPointer = c.isPointer(typName)
	ctx.isArray = c.isArray(typName)
	ctx.extractedType = c.extractBaseType(typName)
}

func (c *Checker) checkFuncDef(stmt *FuncDefStmt) {
	c.enterFuncScope(stmt.Name.Name)
	defer c.CloseScope()
	sym := c.LookupSym(stmt.Name.Name)
	Assert(sym.IsFuncDef(), "Invalid operation on symbol")
	funcdef := sym.(*FuncDef)
	defer func() {
		funcdef.Completed = true
	}()

	if funcdef.Completed {
		return
	}

	ctx := &FuncCtx{
		parent:        nil,
		sym:           funcdef,
		isProblematic: false,
	}
	stmt.setCtx(ctx)
	c.newCtx(ctx)
	defer c.leaveCtx()

	t := c.extractBaseType(funcdef.Type)
	typ := c.LookupSym(t).(*TypeDef)
	if typ == nil {
		ctx.isProblematic = true
		start, end := funcdef.DefNode.Type.Pos()
		c.errorf(start, end, "Type %s is not defined", funcdef.Type)
	} else {
		ctx.expectedType = funcdef.Type
		ctx.returnsStruct = typ.IsStruct
		ctx.returnTypeSym = typ
	}

	for _, param := range funcdef.Params {
		if c.GetVariable(param.Name) == nil {
			c.DefineVar(param.Name, param.Type)
		} else {
			start, end := param.Param.Pos()
			c.errorf(start, end, "re defined param %s", param.Name)
		}

		paramType := c.LookupSym(c.extractBaseType(param.Type))
		if paramType == nil || !paramType.IsTypeDef() {
			start, end := param.Param.Type.Pos()
			c.errorf(start, end, "Provided type %s is not defined", param.Type)
		}
	}

	if funcdef.External {
		return
	}

	c.checkBlockStmt(stmt.Body)
}

func (c *Checker) checkImplStmt(stmt *ImplStmt) {
	ofType := c.getNameOfType(stmt.Type)
	sym := c.LookupSym(ofType)
	if sym == nil || !sym.IsTypeDef() {
		start, end := stmt.Type.Pos()
		c.errorf(start, end, "type %s is not defined", ofType)
		return
	}

	typSym := sym.(*TypeDef)
	stmt.RefOfType = typSym.DefNode
	c.enterTypeScope(typSym.TypeName())
	defer c.CloseScope()

	for _, innerStmt := range stmt.Stmts {
		switch innerStmt := innerStmt.(type) {
		case *FuncDefStmt:
			c.collectFuncSignature(innerStmt, typSym.TypeName())
		default:
			panic("Unreachable impl-block statement")
		}
	}

	for _, innerStmt := range stmt.Stmts {
		switch innerStmt := innerStmt.(type) {
		case *FuncDefStmt:
			c.checkFuncDef(innerStmt)
		default:
			panic("Unreachable impl-block statement")
		}
	}

}

func (c *Checker) checkStmt(stmt Stmt) {
	switch stmt := stmt.(type) {
	case *BlockStmt:
		c.checkBlockStmt(stmt)
	case *VarDefStmt:
		c.checkVarDef(stmt)
	case *ImplStmt:
		c.checkImplStmt(stmt)
	case *ExprStmt:
		c.checkExprStmt(stmt)
	case *ReturnStmt:
		var exprType = "void"
		if stmt.Value != nil {
			var err error
			exprType, err = c.checkExpr(stmt.Value)
			if err != nil {
				return
			}
		}
		stmt.setCtx(c.ctx)
		if !c.checkTypeCompatibility(c.ctx.(*FuncCtx).expectedType, exprType) {
			start, end := stmt.Value.Pos()
			c.errorf(start, end, "expected %s, got %s", c.ctx.(*FuncCtx).expectedType, exprType)
		}
	case *IfStmt:
		l, err := c.checkExpr(stmt.Cond)
		if err != nil {
			return
		}
		if l != "bool" {
			start, end := stmt.Cond.Pos()
			c.errorf(start, end, "condition type is invalid, expected: %s, got: %s", "bool", l)
		}

		c.checkStmt(stmt.Then)
		if stmt.Else != nil {
			c.checkStmt(stmt.Else)
		}

	case *ForStmt:
		// TODO
		c.checkStmt(stmt.Body)
		//Assert(false, "implement me")

	default:
		panic("Unhandled")

	}
}

func isNumber(t string) bool {
	return t == "i8" || t == "i16" || t == "i32" || t == "i64" || t == "f16" || t == "f32" || t == "f64"
}

// checkExpr checks expr and returns its type and error if it exists
func (c *Checker) checkExpr(expr Expr) (string, error) {
	switch expr := expr.(type) {
	case *NumberExpr:
		expr.setCtx(c.ctx)
		if ctx, ok := c.ctx.(ExpectedTypeCtx); ok {
			return "i" + strconv.Itoa(getBitSize(ctx.ExpectedType())), nil
		}
		return "i" + strconv.Itoa(64), nil
	case *FloatExpr:
		expr.setCtx(c.ctx)
		if ctx, ok := c.ctx.(ExpectedTypeCtx); ok {
			return "f" + strconv.Itoa(getBitSize(ctx.ExpectedType())), nil
		}
		return "f" + strconv.Itoa(64), nil
	case *StringExpr:
		expr.setCtx(c.ctx)
		if ctx, ok := c.ctx.(ExpectedTypeCtx); ok {
			if ctx.ExpectedType() == "pointer<char>" {
				return "pointer<char>", nil
			}
		}
		return "string", nil
	case *CharExpr:
		return "char", nil
	case *BoolExpr:
		return "bool", nil
	case *ArrayLitExpr:
		var firstType string
		for i, elem := range expr.Elems {
			elemType, err := c.checkExpr(elem)
			if err != nil {
				continue
			}
			if i == 0 {
				firstType = elemType
			} else {
				if !c.checkTypeCompatibility(firstType, elemType) {
					start, end := elem.Pos()
					c.errorf(start, end, "expected %s, got %s", firstType, elemType)
				}
			}
		}
		return fmt.Sprintf(arrayScheme, firstType, len(expr.Elems)), nil
	case *ObjectLitExpr:
		typ := c.getNameOfType(expr.Type)
		lookupTyp := c.LookupSym(typ)
		if lookupTyp == nil || !lookupTyp.IsTypeDef() {
			start, end := expr.Type.Pos()
			c.errorf(start, end, "provided type %s is not defined", typ)
			return unresolvedType, fmt.Errorf("unknown type %s", typ)
		}

		def := lookupTyp.(*TypeDef)
		typeOfKey := func(name string) (string, error) {
			for _, field := range def.Fields {
				if field.Name == name {
					return field.Type, nil
				}
			}

			return unresolvedType, fmt.Errorf("unknown key")
		}

		for _, kv := range expr.KeyValue {
			k, ok := kv.Key.(*IdentExpr)

			if !ok {
				start, end := k.Pos()
				c.errorf(start, end, "provided key must be ident")
				continue
			}
			l, err := typeOfKey(k.Name)

			if err != nil {
				start, end := k.Pos()
				c.errorf(start, end, err.Error())
				continue
			}

			ctx := &VarAssignCtx{parent: c.ctx, expectedType: l}
			c.newCtx(ctx)
			ctx.arraySize = c.getArraySizeFromTypeStr(l)
			r, err := c.checkExpr(kv.Value)
			if err != nil {
				continue
			}

			if !c.checkTypeCompatibility(l, r) {
				start, end := kv.Pos()
				c.errorf(start, end, "type mismatch (%s:%s)", l, r)
			} else {
				if ctxter, ok := kv.Value.(Contexter); ok {
					ctxter.setCtx(ctx)
				}
				c.setVarAssignCtxFields(ctx, l, c.LookupSym(l).(*TypeDef))
			}

			c.leaveCtx()
		}

		return def.Name, nil
	case *BinaryExpr:

		left, err := c.checkExpr(expr.Left)
		if err != nil {
			return left, err
		}

		right, err := c.checkExpr(expr.Right)
		if err != nil {
			return right, err
		}

		// TODO compability cases

		switch expr.Op {
		case Add, Sub, Mul, Div, Mod:
		case Band:
		case Eq, Neq, Lt, Lte, Gt, Gte, Not, And, Or:
			return "bool", nil
		}

		if left != right {
			start, end := expr.Pos()
			c.errorf(start, end, "types are not compatible %s, %s", left, right)
			return unresolvedType, fmt.Errorf("types are not compatible %s, %s", left, right)
		}

		return left, nil

	case *SelectorExpr:
		left, err := c.checkExpr(expr.Selector)
		if err != nil {
			return unresolvedType, err
		}

		var ltyp = c.extractBaseType(left)

		typDef := c.LookupSym(ltyp)
		if typDef == nil || !typDef.IsTypeDef() {
			start, end := expr.Selector.Pos()
			c.errorf(start, end, "provided type %s is not defined", left)
			return unresolvedType, fmt.Errorf("unknown type %s", left)
		}

		c.enterTypeScope(ltyp)
		defer c.turnTempScopeBack()
		right, err := c.checkExpr(expr.Ident)
		if err != nil {
			return unresolvedType, err
		}

		return right, err

	case *IdentExpr:
		sym := c.LookupSym(expr.Name)
		if sym == nil {
			start, end := expr.Pos()
			c.errorf(start, end, "symbol %s not found", expr.Name)
			return unresolvedType, fmt.Errorf("symbol %s not found", expr.Name)
		}

		c.currentSym = sym
		n := c.extractBaseType(sym.TypeName())
		var typedef = c.LookupSym(n)

		if typedef == nil {
			start, end := expr.Pos()
			c.errorf(start, end, "type is not defined: %s", sym.TypeName())
			return unresolvedType, fmt.Errorf("type is not defined: %s", sym.TypeName())
		}

		return sym.TypeName(), nil
	case *NilExpr:
		return "pointer<nil>", nil
	case *IndexExpr:
		l, err := c.checkExpr(expr.Left)
		if err != nil {
			return unresolvedType, err
		}
		c.newCtx(&IndexCtx{parent: c.ctx, expectedType: "i64"})
		defer c.leaveCtx()
		r, err := c.checkExpr(expr.Index)
		if err != nil {
			// TODO maybe return lefthand type but it is ok for now
			return unresolvedType, err
		}

		// TODO add map in here when we implement that
		if !c.isArray(l) && l != "string" && !c.isPointer(l) {
			start, end := expr.Pos()
			c.errorf(start, end, "invalid access expression either left hand side must be array or map but got: %s", l)
			return unresolvedType, fmt.Errorf("invalid access expression either left hand side must be array or map but got: %s", l)
		}

		if r != "i64" {
			start, end := expr.Index.Pos()
			c.errorf(start, end, "invalid index access expression, expected %s, got %s", "i64", r)
			return unresolvedType, fmt.Errorf("invalid index access expression, expected %s, got %s", "i64", r)
		}

		return c.extractType(l), nil

	case *UnaryExpr:
		switch expr.Op {
		case Add, Sub:
			right, err := c.checkExpr(expr.Right)
			if err != nil {
				return unresolvedType, err
			}

			if right != "i32" {
				start, end := expr.Pos()
				c.errorf(start, end, "right operand must be int")
				return unresolvedType, fmt.Errorf("right operand must be int")
			}

			return "i32", nil
		case Not:
			right, err := c.checkExpr(expr.Right)
			if err != nil {
				return unresolvedType, err
			}
			if right != "bool" {
				start, end := expr.Pos()
				c.errorf(start, end, "right operand must be bool")
				return unresolvedType, fmt.Errorf("right operand must be bool")
			}
			return "bool", nil
		case Mul:
			right, err := c.checkExpr(expr.Right)
			if err != nil {
				return unresolvedType, err
			}

			// it is dereference operation and right operand must be pointer
			if !strings.HasPrefix(right, "pointer<") {
				start, end := expr.Right.Pos()
				c.errorf(start, end, "right operand must be pointer")
				return unresolvedType, fmt.Errorf("right operand must be pointer")
			}

			r := c.extractBaseType(right)
			Assert(c.isPointer(right), "dereferenced type is not pointer")
			return r, nil
		case Band:
			right, err := c.checkExpr(expr.Right)
			if err != nil {
				return unresolvedType, err
			}

			sym := c.LookupSym(right)
			if sym == nil || !sym.IsTypeDef() {
				start, end := expr.Right.Pos()
				c.errorf(start, end, "provided type is not valid: %s", right)
				return unresolvedType, fmt.Errorf("provided type is not valid: %s", right)
			}
			return fmt.Sprintf(pointerScheme, right), nil
		case Sizeof:
			switch r := expr.Right.(type) {
			case *IdentExpr:
				var typName = r.Name
				n := c.extractBaseType(r.Name)
				var sym = c.LookupSym(n)
				if sym == nil || !sym.IsTypeDef() {
					start, end := expr.Right.Pos()
					c.errorf(start, end, "type is not found: %s", typName)
					return unresolvedType, fmt.Errorf("type is not found: %s", typName)
				}
				return "i64", nil
			default:
				start, end := expr.Right.Pos()
				c.errorf(start, end, "unexpected token, expected type identifier")
				return unresolvedType, errors.New("unexpected token, expected type identifier")
			}
		default:
			panic("unreachable")
		}
	case *AssignExpr:
		l, err := c.checkExpr(expr.Left)

		if err != nil {
			return unresolvedType, err
		}

		ctx := &VarAssignCtx{
			parent:       c.ctx,
			expectedType: l,
		}

		c.newCtx(ctx)
		expr.setCtx(ctx)
		defer c.leaveCtx()

		r, err := c.checkExpr(expr.Right)
		if err != nil {
			return unresolvedType, err
		}

		if !c.checkTypeCompatibility(l, r) {
			start, end := expr.Right.Pos()
			c.errorf(start, end, "invalid assignment: expected %s, got %s", l, r)
			return unresolvedType, fmt.Errorf("invalid assignment: expected %s, got: %s", l, r)
		} else {
			sym := c.LookupSym(c.extractBaseType(r))
			if sym.IsTypeDef() {
				c.setVarAssignCtxFields(ctx, sym.TypeName(), sym.(*TypeDef))
			}
		}

		return l, nil

	case *CallExpr:

		ctx := &CallCtx{parent: c.ctx}
		c.newCtx(ctx)
		defer c.leaveCtx()
		expr.setCtx(ctx)

		l, err := c.checkExpr(expr.Left)

		if err != nil {
			return unresolvedType, err
		}

		if !c.currentSym.IsFuncDef() {
			start, end := expr.Pos()
			c.errorf(start, end, "symbol %s is not a function", l)
			return unresolvedType, fmt.Errorf("symbol %s is not a function", l)
		}

		funcDef := c.currentSym.(*FuncDef)
		expr.TypeCast = funcDef.TypeCast
		expr.MethodOf = funcDef.MethodOf
		if !funcDef.Variadic && len(funcDef.Params) != len(expr.Args) && !funcDef.TypeCast {
			start, end := expr.Pos()
			c.errorf(start, end, "expected %d arguments, got %d", len(funcDef.Params), len(expr.Args))
			err = fmt.Errorf("expected %d arguments, got %d", len(funcDef.Params), len(expr.Args))
		} else if funcDef.Variadic {
			if len(expr.Args) < len(funcDef.Params) {
				start, end := expr.Pos()
				c.errorf(start, end, "expected at least %d arguments, got %d", len(funcDef.Params), len(expr.Args))
				err = fmt.Errorf("expected at least %d arguments, got %d", len(funcDef.Params), len(expr.Args))
			}
		}

		var isMalloc = funcDef.Name == "malloc_internal" && funcDef.MethodOf == ""
		if !funcDef.TypeCast {
			for i, _ := range expr.Args {

				ctx := &VarAssignCtx{parent: c.ctx}
				if i < len(funcDef.Params) {
					c.newCtx(ctx)
					arg := funcDef.Params[i]
					ctx.expectedType = arg.Type
				}

				t, err2 := c.checkExpr(expr.Args[i])
				if err2 != nil {
					if err == nil {
						err = err2
					}
					continue
				}

				if i < len(funcDef.Params) {
					arg := funcDef.Params[i]
					if !c.checkTypeCompatibility(arg.Type, t) {
						start, end := expr.Args[i].Pos()
						c.errorf(start, end, "expected %s, got %s", arg.Type, t)
						err = fmt.Errorf("expected %s, got %s", arg.Type, t)
					} else {
						c.setVarAssignCtxFields(ctx, arg.Type, c.LookupSym(c.extractType(arg.Type)).(*TypeDef))
					}
					c.leaveCtx()
				}
			}
		} else {
			err = c.checkTypeCast(expr, funcDef)
		}

		if isMalloc {
			return fmt.Sprintf(autoCastScheme, l), err
		}

		return l, err
	default:
		panic("unreachable")
	}

}

func (c *Checker) checkTypeCast(expr *CallExpr, funcDef *FuncDef) (err error) {
	if len(expr.Args) != 1 {
		start, end := expr.Pos()
		c.errorf(start, end, "expected 1 argument, got %d", len(expr.Args))
		err = fmt.Errorf("expected 1 argument, got %d", len(expr.Args))
	}

	if len(expr.Args) > 0 {
		t, err2 := c.checkExpr(expr.Args[0])
		if err2 == nil {
			funcType := funcDef.Type
			if isNumber(funcType) && (!isNumber(t) && t != "char") {
				start, end := expr.Args[0].Pos()
				c.errorf(start, end, "expected %s, got %s", "<number_type>", t)
				err = fmt.Errorf("expected %s, got %s", "<number_type>", t)
			}

		}
	}

	return err
}
func (c *Checker) checkExprStmt(stmt *ExprStmt) {
	_, _ = c.checkExpr(stmt.Expr)
}

func (c *Checker) checkBlockStmt(stmt *BlockStmt) {
	for _, innerStmt := range stmt.Stmts {
		c.checkStmt(innerStmt)
	}
}

func (c *Checker) check() {
	for _, stmt := range c.Package.Stmts {
		switch stmt := stmt.(type) {
		case *ImplStmt:
			c.checkImplStmt(stmt)
		case *StructDefStmt:
			c.checkStructDef(stmt)
		case *VarDefStmt:
			c.checkVarDef(stmt)
		case *FuncDefStmt:
			c.checkFuncDef(stmt)
		}
	}
}

func (c *Checker) collectFuncSignature(stmt *FuncDefStmt, methodOf string) {
	if _, ok := c.Symbols[stmt.Name.Name]; ok {
		start, end := stmt.Pos()
		c.errorf(start, end, "redeclared symbol %s", stmt.Name.Name)
		return
	}

	funcDef := &FuncDef{
		DefNode:  stmt,
		Name:     stmt.Name.Name,
		Type:     c.getNameOfType(stmt.Type),
		External: stmt.IsExternal,
		MethodOf: methodOf,
	}

	params := make([]Param, 0)
	if methodOf != "" {
		_ = c.addSymbol("this", &VarDef{
			Name: "this",
			Type: methodOf,
		})
	}

	for _, param := range stmt.Params {
		params = append(params, Param{
			Name:  param.Name.Name,
			Type:  c.getNameOfType(param.Type),
			Param: param,
		})
	}
	funcDef.Params = params
	c.Symbols[stmt.Name.Name] = funcDef
}

func (c *Checker) collectSignatures() {
	for _, stmt := range c.Package.Stmts {
		switch stmt := stmt.(type) {
		case *StructDefStmt:

			if _, ok := c.Symbols[stmt.Name.Name]; ok {
				panic("Not implemented")
			}

			typeDef := &TypeDef{
				DefNode:   stmt,
				Name:      stmt.Name.Name,
				Fields:    make([]TypeField, 0),
				Completed: false,
				Methods:   make([]any, 0),
			}

			for _, field := range stmt.Fields {
				typeDef.Fields = append(typeDef.Fields, TypeField{
					Name: field.Name.Name,
					Type: c.getNameOfType(field.Type),
					Node: field,
				})
			}

			c.Symbols[stmt.Name.Name] = typeDef
		case *VarDefStmt:
			if _, ok := c.Symbols[stmt.Name.Name]; ok {
				panic("Not implemented")
			}
			c.Symbols[stmt.Name.Name] = &VarDef{
				DefNode: stmt,
				Name:    stmt.Name.Name,
				Type:    c.getNameOfType(stmt.Type),
			}
		case *FuncDefStmt:
			c.collectFuncSignature(stmt, "")
		}
	}
}
