package main

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

const arrayScheme = "array<%s, %d>"
const sliceScheme = "slice<%s>"
const pointerScheme = "pointer<%s>"
const unresolvedType = "<unresolved>"
const autoCastScheme = "auto_cast<%s>"
const selectorScheme = "selector<%s, %s>"
const contextScheme = "context_based<%s,...>"

type Symbol interface {
	IsSymbol()
	Pos() (start Pos, end Pos)
	IsVarDef() bool
	IsTypeDef() bool
	IsFuncDef() bool
	IsPackage() bool
	TypeName() string
}

type UseSym struct {
	Def    *UseStmt
	Alias  string
	Name   string
	Path   string
	Module *Module
}

func (u *UseSym) IsSymbol() {}
func (u *UseSym) Pos() (start Pos, end Pos) {
	return u.Def.Pos()
}
func (u *UseSym) IsVarDef() bool  { return false }
func (u *UseSym) IsTypeDef() bool { return false }
func (u *UseSym) IsFuncDef() bool { return false }
func (u *UseSym) TypeName() string {
	return u.Name
}

func (u *UseSym) IsPackage() bool { return true }

type TypeField struct {
	Node       *Field
	Name, Type string
}

type Param struct {
	Param               *ParamExpr
	Name                string
	Type                string
	GenericTypeResolver bool
	Optional            bool
}

type TypeDef struct {
	DefNode      DefStmt
	Name         string
	Fields       []TypeField
	Methods      []*FuncDef
	Completed    bool
	IsStruct     bool
	IsBuiltin    bool
	Package      string
	Scope        *Scope
	ReferredFrom *TypeDef
	ImportedFrom *Module
	PackagePath  string
}

func (t *TypeDef) IsPackage() bool {
	return false
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
	DefNode  *VarDefStmt
	Name     string
	Type     string
	IsGlobal bool
	IsConst  bool
	Package  string
	External bool
}

func (v *VarDef) IsPackage() bool { return false }
func (v *VarDef) IsSymbol()       {}
func (v *VarDef) IsVarDef() bool  { return true }
func (v *VarDef) IsTypeDef() bool { return false }
func (v *VarDef) IsFuncDef() bool { return false }
func (v *VarDef) Pos() (Pos, Pos) {
	if v.DefNode != nil {
		return v.DefNode.Pos()
	}

	return Pos{}, Pos{}
}
func (v *VarDef) TypeName() string {
	return v.Type
}

type FuncDef struct {
	DefNode     *FuncDefStmt
	Name        string
	Type        string // indicates return type
	Params      []Param
	MethodOf    string
	MethodOfTyp *TypeDef
	External    bool
	Variadic    bool
	Completed   bool
	TypeCast    bool
	CheckParams bool
	GenericType string
	Package     string
	PackagePath string
	IsBuiltin   bool
}

func (f *FuncDef) IsPackage() bool { return false }
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

func (c *Checker) enterPackageScope(name string) *Scope {
	c.tmpScope = c.Scope
	if c.packageScopes[name] == nil {
		c.packageScopes[name] = c.EnterScope()
	}
	c.Scope = c.packageScopes[name]
	return c.packageScopes[name]
}

func (c *Checker) enterTypeScope(name string) *Scope {
	c.tmpScope = c.Scope
	if c.typeScopes[name] == nil {
		c.typeScopes[name] = c.EnterScope()
	}
	sym := c.LookupSym(name)
	typSym := sym.(*TypeDef)
	c.Scope = c.typeScopes[name]

	if typSym.Scope == nil {
		typSym.Scope = c.typeScopes[name]
	}
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
	initScope      *Scope
	tmpScope       *Scope
	Package        *Package
	Errors         []Error
	currentFile    string
	typeScopes     map[string]*Scope
	packageScopes  map[string]*Scope
	TypeDefs       []*TypeDef
	FuncDefs       []*FuncDef
	GlobalVarDefs  []*VarDef
	Imports        []*Module
	ImportMap      map[string]*Module  // importPath -> module
	PathAliasMap   map[string]string   // importPath -> alias
	ImportAliasMap map[string][]string // fileName -> aliases

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
	c.initScope = c.Scope
	c.typeScopes = make(map[string]*Scope)
	c.packageScopes = make(map[string]*Scope)
	defer c.CloseScope()
	c.initGlobalScope()
	c.collectSignatures()
	stmts := make([]*VarDefStmt, 0)
	allStmts := c.Package.AllStatements()
	for _, stmt := range allStmts {
		switch stmt := stmt.(type) {
		case *VarDefStmt:
			stmts = append(stmts, stmt)
		}
	}

	if !c.checkInitializationCycle(stmts) {
		c.topologicalSort(stmts)
	}

	sort.SliceStable(c.Package.Stmts, func(i, j int) bool {
		a, b := c.Package.Stmts[i], c.Package.Stmts[j]
		aa, ok := a.(*VarDefStmt)
		bb, ok2 := b.(*VarDefStmt)

		if ok && ok2 {
			return aa.Order < bb.Order
		}

		if ok || ok2 {
			return true
		}

		return false

	})
	c.check()
	return c.Errors, len(c.Errors) > 0
}

func (c *Checker) addSymbol(name string, sym Symbol) error {
	if _, ok := c.Scope.Symbols[name]; ok {
		return fmt.Errorf("duplicate symbol %s", name)
	}
	c.Scope.Symbols[name] = sym

	switch sym := sym.(type) {
	case *VarDef:
		if sym.IsGlobal && !sym.External {
			c.GlobalVarDefs = append(c.GlobalVarDefs, sym)
		}
	case *TypeDef:
		c.TypeDefs = append(c.TypeDefs, sym)
	case *FuncDef:
		c.FuncDefs = append(c.FuncDefs, sym)
	case *UseSym:
		c.Imports = append(c.Imports, sym.Module)
	}

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
			IsBuiltin: true,
			Package:   c.Package.Name,
		})

		if name != "void" {
			_ = c.addSymbol("to_"+name, &FuncDef{
				DefNode:   nil,
				Name:      "to_" + name,
				Type:      name,
				Params:    []Param{},
				MethodOf:  "",
				External:  false,
				Variadic:  false,
				Completed: true,
				TypeCast:  true,
				IsBuiltin: true,
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
		External:    true,
		Variadic:    true,
		Completed:   false,
		CheckParams: true,
		IsBuiltin:   true,
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
		External:    true,
		Variadic:    true,
		Completed:   false,
		CheckParams: true,
		IsBuiltin:   true,
	})
	AssertErr(err)

	err = c.addSymbol("append", &FuncDef{
		Name: "append",
		Type: "void",
		Params: []Param{
			{
				Name:                "list",
				Type:                "slice<!T>",
				GenericTypeResolver: true,
			},
			{
				Name: "value",
				Type: "!T",
			},
		},
		External:    true,
		Completed:   true,
		GenericType: "!T",
		IsBuiltin:   true,
	})
	AssertErr(err)

	err = c.addSymbol("len", &FuncDef{
		Name: "len",
		Type: "i64",
		Params: []Param{
			{
				Name:                "slice",
				Type:                "slice<!T>",
				GenericTypeResolver: true,
			},
		},
		External:    true,
		Completed:   true,
		GenericType: "!T",
		IsBuiltin:   true,
	})
	AssertErr(err)

	err = c.addSymbol("print", &FuncDef{
		Name: "print",
		Type: "void",
		Params: []Param{
			{
				Name: "buffer",
				Type: "context_based<i8,i16,i32,i64,char,bool,string,f16,f32,f64,pointer<char>>",
			},
		},
		MethodOf:    "",
		External:    false,
		Variadic:    false,
		Completed:   false,
		TypeCast:    false,
		CheckParams: false,
		GenericType: "",
		IsBuiltin:   true,
	})
	AssertErr(err)

	err = c.addSymbol("println", &FuncDef{
		Name: "println",
		Type: "void",
		Params: []Param{
			{
				Name:     "buffer",
				Type:     "context_based<i8,i16,i32,i64,char,bool,string,f16,f32,f64,pointer<char>>",
				Optional: true,
			},
		},
		MethodOf:    "",
		External:    false,
		Variadic:    false,
		Completed:   false,
		TypeCast:    false,
		CheckParams: false,
		GenericType: "",
		IsBuiltin:   true,
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

func (c *Checker) LookupSym(name string) Symbol {
	sym := c.Scope.LookupSym(name)
	if sym != nil && sym.IsTypeDef() {
		typSym := sym.(*TypeDef)
		if !typSym.IsBuiltin && typSym.Package != c.Package.Name {
			extractedSymName := strings.Split(typSym.Name, ".")[0]
			if !c.checkImported(extractedSymName) {
				return nil
			}
		}
	}
	return sym
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
		Package: c.Package.Name,
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
			return fmt.Sprintf(arrayScheme, c.getNameOfType(expr.Type), size)
		} else {
			expr.Size = &NumberExpr{Value: -1}
			return fmt.Sprintf(sliceScheme, c.getNameOfType(expr.Type))
		}
	case *SelectorExpr:
		left := c.getNameOfType(expr.Selector)
		return fmt.Sprintf("%s.%s", left, c.getNameOfType(expr.Ident))
	default:
		panic("unreachable")
	}
}

func (c *Checker) errorf(start, end Pos, format string, args ...interface{}) {
	c.Errors = append(c.Errors, Error{
		Start:   start,
		End:     end,
		Message: fmt.Sprintf(format, args...),
		File:    c.currentFile,
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

		if fieldType != nil && fieldType.TypeName() == stmt.Name.Name && !c.isPointer(field.Type) && !c.isSlice(field.Type) {
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
	} else if strings.HasPrefix(name, "context_based<") {
		if c.ctx == nil {
			return unresolvedType
		}

		if _, ok := c.ctx.(ExpectedTypeCtx); !ok {
			return unresolvedType
		}

		var s = strings.Replace(name, "context_based<", "", 1)
		s = strings.Replace(s, ">", "", 1)
		var typeList []string
		typeList = strings.Split(s, ",")
		var first string
		for i, typ := range typeList {
			typ = strings.TrimSpace(typ)
			if i == 0 {
				first = typ
			}
			if c.ctx.(ExpectedTypeCtx).ExpectedType() == typ {
				return typ
			}
		}

		return first
	} else if strings.HasPrefix(name, "slice<") {
		var s string
		fmt.Sscanf(name, sliceScheme, &s)
		s = strings.TrimSuffix(s, ">")
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

func (c *Checker) isSlice(name string) bool {
	return strings.HasPrefix(name, "slice<")
}

func (c *Checker) isContextBased(name string) bool {
	return strings.HasPrefix(name, "context_based<")
}

func (c *Checker) extractContextBasedTypes(name string) []string {
	name = strings.Replace(name, "context_based<", "", 1)
	name = strings.Replace(name, ">", "", 1)
	return strings.Split(name, ",")
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

	if c.isSlice(typ1) && c.isSlice(exprType) {
		typ1Base := c.extractBaseType(typ1)
		exprTypeBase := c.extractBaseType(exprType)
		if typ1Base != exprTypeBase {
			return false
		}
	}

	if c.isContextBased(typ1) {
		possibleTypes := c.extractContextBasedTypes(typ1)
		for _, t := range possibleTypes {
			if t == exprType {
				return true
			}
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
	} else if typName == "i8" || typName == "char" {
		return 8
	} else if typName == "f64" || typName == "i64" {
		return 64
	}

	return 64
}

// extractBaseType extracts type until base type
func (c *Checker) extractBaseType(name string) string {

	if strings.HasPrefix(name, "pointer<") {
		return c.extractBaseType(c.extractType(name))
	} else if strings.HasPrefix(name, "auto_cast<") {
		return c.extractType(c.extractType(name))
	} else if strings.HasPrefix(name, "array<") {
		return c.extractType(c.extractType(name))
	} else if strings.HasPrefix(name, "slice<") {
		return c.extractType(c.extractType(name))
	}

	// it is already unwrapped so
	return name
}

func (c *Checker) fixVarDefStmt(stmt *VarDefStmt, l, r string, ctx *VarAssignCtx) {
	if c.isArray(l) && c.isArray(r) {
		if c.getArraySizeFromTypeStr(l) == -1 {
			ctx.arraySize = -1
			stmt.Type.(*ArrayTypeExpr).Size = &NumberExpr{Value: -1}
		} else {
			ctx.arraySize = c.getArraySizeFromTypeStr(r)
		}

		if alit, ok := stmt.Init.(*ArrayLitExpr); ok {
			alit.setCtx(ctx)
		}
	} else if c.isSlice(l) && c.isSlice(r) {
		ctx.arraySize = -1
		ctx.isSlice = true
		stmt.Type.(*ArrayTypeExpr).Size = &NumberExpr{Value: -1}
		if alit, ok := stmt.Init.(*ArrayLitExpr); ok {
			alit.setCtx(ctx)
		}
	}
}
func (c *Checker) checkVarDef(stmt *VarDefStmt, isGlobal bool) {
	variable := c.GetVariable(stmt.Name.Name)
	if variable != nil && !isGlobal {
		start, end := stmt.Name.Pos()
		c.errorf(start, end, "redeclared symbol %s", stmt.Name.Name)
	}

	if isGlobal {
		if !variable.IsVarDef() {
			start, end := stmt.Name.Pos()
			c.errorf(start, end, "already declared as non variable: %s", stmt.Name.Name)
		}
	}

	typName := c.getNameOfType(stmt.Type)
	var typ = c.extractBaseType(typName)
	typSym := c.LookupSym(typ)
	ctx := &VarAssignCtx{
		parent:       c.ctx,
		expectedType: typName,
		isSlice:      c.isSlice(typName),
		isGlobal:     isGlobal,
		isConstant:   stmt.IsConst,
	}

	c.newCtx(ctx)
	defer c.leaveCtx()
	stmt.setCtx(ctx)

	if typSym == nil || !typSym.IsTypeDef() {
		start, end := stmt.Type.Pos()
		c.errorf(start, end, "type %s is not defined", typName)
	}

	if stmt.Init == nil && stmt.IsConst {
		start, end := stmt.Pos()
		c.errorf(start, end, "constants must have initializer")
	}

	if stmt.Init != nil {

		if stmt.IsConst {
			expr, isConstExpr := stmt.Init.(Constant)
			if !isConstExpr || !expr.IsConstant() {
				start, end := stmt.Init.Pos()
				c.errorf(start, end, "expected const initializer")
			}
		}

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
	if variable != nil {
		variable.(*VarDef).IsGlobal = isGlobal
	}

	if !isGlobal {
		c.DefineVar(stmt.Name.Name, typName)
		c.GetVariable(stmt.Name.Name).(*VarDef).IsConst = stmt.IsConst
	}
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
	typ := c.LookupSym(t)
	if typ == nil {
		ctx.isProblematic = true
		ctx.expectedType = funcdef.Type
		start, end := funcdef.DefNode.Type.Pos()
		c.errorf(start, end, "Type %s is not defined", funcdef.Type)
	} else {
		typ := typ.(*TypeDef)
		ctx.expectedType = funcdef.Type
		ctx.returnsStruct = typ.IsStruct && !c.isPointer(funcdef.Type)
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
			c.collectFuncSignature(innerStmt, typSym.TypeName(), typSym)
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

func (c *Checker) checkIncrStmt(stmt *IncrStmt) {
	typ, _ := c.checkExpr(stmt.Expr)
	if !isNumber(typ) {
		start, end := stmt.Pos()
		c.errorf(start, end, "incr(++) stmt only valid on numeric expressions")
	}
	stmt.Type = typ
}

func (c *Checker) checkDecrStmt(stmt *DecrStmt) {
	typ, _ := c.checkExpr(stmt.Expr)
	if !isNumber(typ) {
		start, end := stmt.Pos()
		c.errorf(start, end, "decr(--) stmt only valid on numeric expressions")
	}
	stmt.Type = typ
}

func (c *Checker) checkStmt(stmt Stmt) {
	switch stmt := stmt.(type) {
	case *BlockStmt:
		c.checkBlockStmt(stmt)
	case *VarDefStmt:
		c.checkVarDef(stmt, false)
	case *ImplStmt:
		c.checkImplStmt(stmt)
	case *ExprStmt:
		c.checkExprStmt(stmt)
	case *IncrStmt:
		c.checkIncrStmt(stmt)
	case *DecrStmt:
		c.checkDecrStmt(stmt)
	case *ReturnStmt:
		stmt.setCtx(c.ctx)
		var exprType = "void"
		if stmt.Value != nil {
			var err error
			exprType, err = c.checkExpr(stmt.Value)
			if err != nil {
				return
			}
		}
		if !c.checkTypeCompatibility(stmt.Context().expectedType, exprType) {
			var start, end Pos
			if stmt.Value != nil {
				start, end = stmt.Value.Pos()
			} else {
				start, end = stmt.Pos()
			}
			c.errorf(start, end, "expected %s, got %s", getFuncCtx(c.ctx).expectedType, exprType)
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

		c.EnterScope()
		c.checkStmt(stmt.Then)
		c.CloseScope()
		if stmt.Else != nil {
			c.EnterScope()
			c.checkStmt(stmt.Else)
			c.CloseScope()
		}

	case *ForStmt:
		c.newCtx(&ForCtx{parent: c.ctx})
		defer c.leaveCtx()
		c.EnterScope()
		defer c.CloseScope()
		if stmt.Init != nil {
			c.checkStmt(stmt.Init)
		}
		stmt.ctx = c.ctx.(*ForCtx)

		if stmt.Cond != nil {
			cond, err := c.checkExpr(stmt.Cond)
			if err != nil {
				return
			}

			if cond != "bool" {
				return
			}
		}

		if stmt.Step != nil {
			c.checkStmt(stmt.Step)
		}

		c.checkStmt(stmt.Body)
	//Assert(false, "implement me")
	case *BreakStmt:
		if c.ctx == nil {
			start, end := stmt.Pos()
			c.errorf(start, end, "invalid break statement")
			return
		}

		if _, ok := c.ctx.(*ForCtx); !ok {
			start, end := stmt.Pos()
			c.errorf(start, end, "invalid break statement")
			return
		}
		stmt.ctx = c.ctx.(*ForCtx)
	case *ContinueStmt:
		if c.ctx == nil {
			start, end := stmt.Pos()
			c.errorf(start, end, "invalid continue statement")
			return
		}

		if _, ok := c.ctx.(*ForCtx); !ok {
			start, end := stmt.Pos()
			c.errorf(start, end, "invalid continue statement")
			return
		}
		stmt.ctx = c.ctx.(*ForCtx)
	default:
		panic("Unhandled")

	}
}

func isNumber(t string) bool {
	return t == "i8" || t == "i16" || t == "i32" || t == "i64" || t == "f16" || t == "f32" || t == "f64"
}

type TypeInfo struct {
	child *TypeInfo
	name  string
	typ   string
	size  int
}

func (c *Checker) typeTree(t string) *TypeInfo {

	if c.isSlice(t) {
		t1 := c.extractType(t)
		return &TypeInfo{typ: "slice", name: t, child: c.typeTree(t1)}
	} else if c.isArray(t) {
		extractType := c.extractType(t)
		return &TypeInfo{typ: "array", name: t, child: c.typeTree(extractType), size: c.getArraySizeFromTypeStr(t)}
	} else if c.isPointer(t) {
		s := c.extractType(t)
		return &TypeInfo{typ: "pointer", name: t, child: c.typeTree(s)}
	}

	return &TypeInfo{typ: "base", name: t}
}

func findDepth(tree *TypeInfo, genericType string, i int) int {
	if tree.name == genericType {
		return i
	}

	if tree.child == nil {
		return -1
	}

	return findDepth(tree.child, genericType, i+1)
}

func findDepthValue(tree *TypeInfo, depth int) *TypeInfo {

	if tree == nil {
		return nil
	}

	if depth == 0 {
		return tree
	}

	if tree.child != nil {
		return findDepthValue(tree.child, depth-1)
	}

	return nil
}

func (c *Checker) resolveGenericArgType(funcDef *FuncDef, param Param, right string) string {

	genericType := funcDef.GenericType
	var resolvedTree *TypeInfo

	if param.GenericTypeResolver {
		paramTypTree := c.typeTree(param.Type)
		depth := findDepth(paramTypTree, genericType, 0)

		rightTree := c.typeTree(right)
		resolvedTree = findDepthValue(rightTree, depth)
	}

	if resolvedTree == nil {
		return genericType
	}

	return resolvedTree.name
}

// checkExpr checks expr and returns its type and error if it exists
func (c *Checker) checkExpr(expr Expr) (string, error) {
	switch expr := expr.(type) {
	case *NumberExpr:
		expr.setCtx(c.ctx)
		if ctx, ok := c.ctx.(ExpectedTypeCtx); ok {
			return "i" + strconv.Itoa(getBitSize(c.extractBaseType(ctx.ExpectedType()))), nil
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
		expr.setCtx(c.ctx)
		var firstType string
		ctx := expr.ctx.(*VarAssignCtx)
		if expr.ctx.(*VarAssignCtx).isSlice {
			firstType = c.extractType(ctx.expectedType)
		}
		for _, elem := range expr.Elems {
			elemType, err := c.checkExpr(elem)
			if err != nil {
				continue
			}
			if firstType == "" {
				firstType = elemType
			} else {
				if !c.checkTypeCompatibility(firstType, elemType) {
					start, end := elem.Pos()
					c.errorf(start, end, "expected %s, got %s", firstType, elemType)
				}
			}
		}

		if ctx.isSlice {
			return fmt.Sprintf(sliceScheme, firstType), nil
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

			ctx := &VarAssignCtx{parent: c.ctx, expectedType: l, extractedType: c.extractBaseType(l), isSlice: c.isSlice(l)}
			c.newCtx(ctx)
			ctx.arraySize = c.getArraySizeFromTypeStr(l)
			r, err := c.checkExpr(kv.Value)
			if err != nil {
				c.leaveCtx()
				continue
			}

			if !c.checkTypeCompatibility(l, r) {
				start, end := kv.Pos()
				c.errorf(start, end, "type mismatch (%s:%s)", l, r)
			} else {
				/*if ctxter, ok := kv.Value.(Contexter); ok {
					ctxter.setCtx(ctx)
				}*/
				kv.setCtx(ctx)
				c.setVarAssignCtxFields(ctx, l, c.LookupSym(c.extractBaseType(l)).(*TypeDef))
			}

			c.leaveCtx()
		}

		return typ, nil
	case *BinaryExpr:

		left, err := c.checkExpr(expr.Left)
		if err != nil {
			return left, err
		}

		right, err := c.checkExpr(expr.Right)
		if err != nil {
			return right, err
		}

		expr.Ctx = &BinaryExprCtx{
			parent:    c.ctx,
			IsRuntime: (left == "string" && right == "string") || (left == "string" && right == "char") || (left == "char" && right == "string") || (left == "char" && right == "char" && (expr.Op != Eq && expr.Op != Neq) && c.ctx.(ExpectedTypeCtx).ExpectedType() == "string"),
			OpInfo:    operationInfo{left, expr.Op, right},
		}

		if v, ok := validOps[operationInfo{left, expr.Op, right}]; ok {
			v = c.extractType(v)
			expr.Ctx.ResultType = v
			return v, nil
		} else {
			if v2, ok := validOps[operationInfo{"*", expr.Op, "*"}]; ok {
				v2 = c.extractType(v2)
				expr.Ctx.ResultType = v2
				return v2, nil
			}
			start, end := expr.Pos()
			c.errorf(start, end, "invalid operation: %s %s %s", left, expr.Op, right)
			return unresolvedType, fmt.Errorf("invalid operation: %s %s, %s", left, expr.Op, right)
		}
	case *SelectorExpr:
		left, err := c.checkExpr(expr.Selector)
		if err != nil {
			return unresolvedType, err
		}

		ctx := &SelectorCtx{
			parent:    c.ctx,
			IsPackage: false,
		}
		expr.Ctx = ctx
		var ltyp = c.extractBaseType(left)

		relatedSym := c.LookupSym(ltyp)
		if relatedSym == nil || (!relatedSym.IsTypeDef() && !relatedSym.IsPackage()) {
			start, end := expr.Selector.Pos()
			c.errorf(start, end, "provided type %s is not defined", left)
			return unresolvedType, fmt.Errorf("unknown type %s", left)
		}

		if relatedSym.IsTypeDef() {
			c.enterTypeScope(ltyp)
			defer c.turnTempScopeBack()
		} else {
			ctx.IsPackage = true
			c.enterPackageScope(ltyp)
			defer c.turnTempScopeBack()
		}

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
		c.newCtx(&IndexCtx{parent: c.ctx, expectedType: "i64", sourceBaseType: c.extractBaseType(l), isPointer: c.isPointer(c.extractType(l))})
		expr.ctx = c.ctx.(*IndexCtx)
		defer c.leaveCtx()
		r, err := c.checkExpr(expr.Index)
		if err != nil {
			// TODO maybe return lefthand type but it is ok for now
			return unresolvedType, err
		}

		// TODO add map in here when we implement that
		if !c.isArray(l) && l != "string" && !c.isPointer(l) && !c.isSlice(l) {
			start, end := expr.Pos()
			c.errorf(start, end, "invalid access expression either left hand side must be array or map but got: %s", l)
			return unresolvedType, fmt.Errorf("invalid access expression either left hand side must be array or map but got: %s", l)
		}

		if r != "i64" && r != "i32" && r != "i16" && r != "i8" {
			start, end := expr.Index.Pos()
			c.errorf(start, end, "invalid index access expression, expected %s, got %s", "i64", r)
			return unresolvedType, fmt.Errorf("invalid index access expression, expected %s, got %s", "i64", r)
		}

		if l == "string" {
			return "char", nil
		}

		return c.extractType(l), nil

	case *UnaryExpr:
		switch expr.Op {
		case Add, Sub:
			right, err := c.checkExpr(expr.Right)
			if err != nil {
				return unresolvedType, err
			}

			return right, nil
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

			extracted := c.extractBaseType(right)
			sym := c.LookupSym(extracted)

			if sym == nil || !sym.IsTypeDef() {
				start, end := expr.Right.Pos()
				c.errorf(start, end, "provided type is not valid: %s", extracted)
				return unresolvedType, fmt.Errorf("provided type is not valid: %s", extracted)
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
		case New:
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
				var res = strings.Replace(pointerScheme, "%s", typName, 1)
				return res, nil
			default:
				start, end := expr.Right.Pos()
				c.errorf(start, end, "unexpected token, expected type identifier")
				return unresolvedType, errors.New("unexpected token, expected type identifier")
			}
		default:
			panic("unreachable")
		}
	case *AssignExpr:
		c.checkConstAssign(expr.Left)
		l, err := c.checkExpr(expr.Left)

		if err != nil {
			return unresolvedType, err
		}

		ctx := &VarAssignCtx{
			parent:        c.ctx,
			expectedType:  l,
			extractedType: c.extractBaseType(l),
			isSlice:       c.isSlice(l),
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
			if sym.IsTypeDef() || sym.IsPackage() {
				c.setVarAssignCtxFields(ctx, l, sym.(*TypeDef))
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
		expr.Package = funcDef.Package

		var optionalCount = 0
		for _, p := range funcDef.Params {
			if p.Optional {
				optionalCount++
			}
		}

		if !funcDef.Variadic && (len(funcDef.Params) != len(expr.Args) && len(funcDef.Params)-optionalCount != len(expr.Args)) && !funcDef.TypeCast {

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

		// TODO come up with the cleanest solution
		ctx.isCustomCall = (funcDef.Name == "append" || funcDef.Name == "print" || funcDef.Name == "println") && funcDef.MethodOf == ""
		var resolvedGenericType string
		if !funcDef.TypeCast {
			for i, _ := range expr.Args {

				ctx := &VarAssignCtx{parent: c.ctx}
				if i < len(funcDef.Params) {
					c.newCtx(ctx)
					arg := funcDef.Params[i]
					var argType = arg.Type
					if resolvedGenericType != "" {
						argType = strings.Replace(argType, funcDef.GenericType, resolvedGenericType, 1)
					}
					ctx.expectedType = argType
				}

				t, err2 := c.checkExpr(expr.Args[i])
				if err2 != nil {
					if err == nil {
						err = err2
					}
					if i < len(funcDef.Params) {
						c.leaveCtx()
					}
					continue
				}

				if i < len(funcDef.Params) {
					arg := funcDef.Params[i]
					argType := arg.Type
					if arg.GenericTypeResolver {
						resolvedGenericType = c.resolveGenericArgType(funcDef, arg, t)
					}

					if resolvedGenericType != "" {
						argType = strings.Replace(argType, funcDef.GenericType, resolvedGenericType, 1)
					}

					if !c.checkTypeCompatibility(argType, t) {
						start, end := expr.Args[i].Pos()
						c.errorf(start, end, "expected %s, got %s", argType, t)
						err = fmt.Errorf("expected %s, got %s", argType, t)
					} else {
						var ctxTyp = argType
						if c.isContextBased(argType) {
							ctxTyp = t
						}
						c.setVarAssignCtxFields(ctx, argType, c.LookupSym(c.extractBaseType(ctxTyp)).(*TypeDef))
					}
					c.leaveCtx()
				}
			}
		} else {
			err = c.checkTypeCast(expr, funcDef)
		}

		return l, err
	case *ParenExpr:
		return c.checkExpr(expr.Expr)

	default:
		panic("unreachable")
	}

}

func (c *Checker) restoreExternalTyp(alias, typ, packagePath string) string {
	t := c.extractBaseType(typ)
	t2 := strings.Split(t, ".")
	if len(t2) > 1 {
		alias2 := c.PathAliasMap[packagePath]
		return strings.Replace(typ, t2[0], alias2, 1)
	}

	return strings.Replace(typ, t, fmt.Sprintf("%s.%s", alias, t), 1)
}

func (c *Checker) checkConstAssign(expr Expr) {

	switch expr := expr.(type) {
	case *IdentExpr:
		ident := expr
		v := c.LookupSym(ident.Name)
		if v != nil {
			if v.IsVarDef() {
				if v.(*VarDef).IsConst {
					start, end := ident.Pos()
					c.errorf(start, end, "const variables cannot be changed")
				}
			} else {
				start, end := expr.Pos()
				c.errorf(start, end, "expected identifier for assign expr")
			}
		}
	case *SelectorExpr:
		l, err := c.checkExpr(expr.Selector)
		if err != nil {
			return
		}
		sym := c.LookupSym(l)
		if sym == nil || !sym.IsPackage() {
			return
		}
		c.enterPackageScope(l)
		c.checkConstAssign(expr.Ident)
		c.turnTempScopeBack()
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
			c.ctx.(*CallCtx).typeCastParamType = t
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
		c.currentFile = c.Package.FileMap[stmt]
		switch stmt := stmt.(type) {
		case *UseStmt:
			c.checkUseStmt(stmt)
		}
	}

	for _, stmt := range c.Package.AllStatements() {
		c.currentFile = c.Package.FileMap[stmt]
		switch stmt := stmt.(type) {
		case *ImplStmt:
			c.checkImplStmt(stmt)
		case *StructDefStmt:
			c.checkStructDef(stmt)
		case *VarDefStmt:
			c.checkVarDef(stmt, true)
		case *FuncDefStmt:
			c.checkFuncDef(stmt)
		}
	}
}

func (c *Checker) addImportAlias(path, symName string) {
	Assert(c.currentFile != "", "currentFile cannot be empty!")
	if c.ImportAliasMap[c.currentFile] == nil {
		c.ImportAliasMap[c.currentFile] = make([]string, 0)
	}
	c.ImportAliasMap[c.currentFile] = append(c.ImportAliasMap[c.currentFile], symName)
	c.PathAliasMap[path] = symName
}

func (c *Checker) checkImported(symName string) bool {
	imports := c.ImportAliasMap[c.currentFile]
	if imports == nil {
		return false
	}

	for _, sym := range imports {
		if sym == symName {
			return true
		}
	}

	return false
}

func (c *Checker) checkAlreadyImported(useStmt *UseStmt, symName string) {
	if c.checkImported(symName) {
		start, end := useStmt.Pos()
		c.errorf(start, end, "%s already imported", symName)
	}
}

func (c *Checker) checkUseStmt(stmt *UseStmt) {
	joinPath := path.Join(BasePath, stmt.Path.Raw)
	var module *Module
	if _, ok := c.ImportMap[joinPath]; ok {
		module = c.ImportMap[joinPath]
	} else {
		module = Import(joinPath, c.ImportMap)
		c.ImportMap[joinPath] = module
	}
	stmt.useCtx = &UseCtx{parent: c.ctx, Module: module}
	var symName string
	var alias string
	if stmt.Alias != nil {
		symName = stmt.Alias.Name
		alias = stmt.Alias.Name
	} else {
		pathSplit := strings.Split(joinPath, "/")
		lenSplit := len(pathSplit)
		symName = pathSplit[lenSplit-1]
	}
	stmt.useCtx.Alias = symName
	c.checkAlreadyImported(stmt, symName)
	c.addImportAlias(joinPath, symName)
	for _, typDef := range module.TypeDefs {
		if !typDef.IsBuiltin {
			cloned := *typDef
			cloned.ImportedFrom = module
			cloned.ReferredFrom = typDef
			cloned.Name = symName + "." + typDef.Name
			c.addSymbol(symName+"."+typDef.Name, &cloned)
			c.typeScopes[symName+"."+typDef.Name] = typDef.Scope
		}
	}

	c.enterPackageScope(symName)
	for _, funcDef := range module.FuncDef {
		if !funcDef.IsBuiltin && funcDef.MethodOf == "" {
			cloned := *funcDef
			sym := stmt.useCtx.Module.Scope.LookupSym(c.extractBaseType(cloned.Type))
			if !sym.(*TypeDef).IsBuiltin {
				cloned.Type = c.restoreExternalTyp(symName, cloned.Type, sym.(*TypeDef).PackagePath)
			}

			cloned.Params = make([]Param, 0)
			for _, p := range funcDef.Params {
				cloned.Params = append(cloned.Params, p)
			}

			for i, param := range cloned.Params {
				sym := stmt.useCtx.Module.Scope.LookupSym(c.extractBaseType(param.Type))
				if !sym.(*TypeDef).IsBuiltin {
					param.Type = c.restoreExternalTyp(symName, param.Type, sym.(*TypeDef).PackagePath)
					cloned.Params[i] = param
				}
			}
			AssertErr(c.addSymbol(funcDef.Name, &cloned))
		}
	}

	for _, varDef := range module.Globals {
		cloned := *varDef
		sym := stmt.useCtx.Module.Scope.LookupSym(c.extractBaseType(cloned.Type))
		if !sym.(*TypeDef).IsBuiltin {
			cloned.Type = c.restoreExternalTyp(symName, cloned.Type, sym.(*TypeDef).PackagePath)
		}
		cloned.External = true
		AssertErr(c.addSymbol(varDef.Name, &cloned))
	}

	c.turnTempScopeBack()
	_ = c.addSymbol(symName, &UseSym{Def: stmt, Name: symName, Alias: alias, Path: joinPath, Module: module})

}

func (c *Checker) collectFuncSignature(stmt *FuncDefStmt, methodOf string, typSym *TypeDef) {
	if _, ok := c.Symbols[stmt.Name.Name]; ok {
		start, end := stmt.Pos()
		c.errorf(start, end, "redeclared symbol %s", stmt.Name.Name)
		return
	}

	funcDef := &FuncDef{
		DefNode:     stmt,
		Name:        stmt.Name.Name,
		Type:        c.getNameOfType(stmt.Type),
		External:    stmt.IsExternal,
		MethodOf:    methodOf,
		Package:     c.Package.Name,
		PackagePath: c.Package.Path,
	}

	params := make([]Param, 0)
	if methodOf != "" {
		_ = c.addSymbol("this", &VarDef{
			Name:    "this",
			Type:    fmt.Sprintf(pointerScheme, methodOf),
			Package: c.Package.Name,
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
	err := c.addSymbol(stmt.Name.Name, funcDef)
	AssertErr(err)
	if typSym != nil {
		typSym.Methods = append(typSym.Methods, funcDef)
		funcDef.MethodOfTyp = typSym
	}
}

func (c *Checker) collectSignatures() {
	for _, file := range c.Package.Files {
		for _, stmt := range file.Stmts {
			switch stmt := stmt.(type) {
			case *StructDefStmt:

				if _, ok := c.Symbols[stmt.Name.Name]; ok {
					panic("Not implemented")
				}

				typeDef := &TypeDef{
					DefNode:     stmt,
					Name:        stmt.Name.Name,
					Fields:      make([]TypeField, 0),
					Completed:   false,
					Methods:     make([]*FuncDef, 0),
					Package:     c.Package.Name,
					PackagePath: c.Package.Path,
				}

				for _, field := range stmt.Fields {
					typeDef.Fields = append(typeDef.Fields, TypeField{
						Name: field.Name.Name,
						Type: c.getNameOfType(field.Type),
						Node: field,
					})
				}

				err := c.addSymbol(stmt.Name.Name, typeDef)
				AssertErr(err)
			case *VarDefStmt:
				if _, ok := c.Symbols[stmt.Name.Name]; ok {
					panic("Not implemented")
				}
				err := c.addSymbol(stmt.Name.Name, &VarDef{
					DefNode:  stmt,
					Name:     stmt.Name.Name,
					Type:     c.getNameOfType(stmt.Type),
					IsConst:  stmt.IsConst,
					IsGlobal: true,
					Package:  c.Package.Name,
				})
				AssertErr(err)
			case *FuncDefStmt:
				c.collectFuncSignature(stmt, "", nil)
			}
		}
	}
}
