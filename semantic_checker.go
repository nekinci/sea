package main

import (
	"fmt"
	"strconv"
	"strings"
)

const arrayScheme = "array<%s, %d>"
const pointerScheme = "pointer<%s>"
const unresolvedType = "<unresolved>"

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
	expectedFuncType string
	currentSym       Symbol
	expectedBitSize  []int
	//
}

func (c *Checker) pushExpectedBitSize(i int) {
	c.expectedBitSize = append(c.expectedBitSize, i)
}

func (c *Checker) popExpectedBitSize() int {
	if len(c.expectedBitSize) == 0 {
		panic("empty expectedBitSize")
	}

	i := c.expectedBitSize[len(c.expectedBitSize)-1]
	c.expectedBitSize = c.expectedBitSize[:len(c.expectedBitSize)-1]
	return i
}

func (c *Checker) topExpectedBitSize() int {
	if len(c.expectedBitSize) == 0 {
		panic("empty expectedBitSize")
	}
	return c.expectedBitSize[len(c.expectedBitSize)-1]
}

func (c *Checker) dupExpectedBitSize() int {
	b := c.expectedBitSize[len(c.expectedBitSize)-1]
	c.pushExpectedBitSize(b)
	return b
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

	_ = c.addSymbol("printf_internal", &FuncDef{
		DefNode: nil,
		Name:    "printf_internal",
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

	_ = c.addSymbol("scanf_internal", &FuncDef{
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
	if sym := c.LookupSym(name); sym != nil {
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
	default:
		panic("unreachable")
	}
}

func (c *Checker) errorf(start, end Pos, format string, args ...interface{}) {
	c.Errors = append(c.Errors, Error{
		Start:   start,
		End:     end,
		Message: fmt.Sprintf(format, args...),
		File:    "./input.file",
	})
}

func (c *Checker) checkStructDef(stmt *StructDefStmt) {
	c.enterTypeScope(stmt.Name.Name)
	defer c.CloseScope()
	sym := c.LookupSym(stmt.Name.Name)
	Assert(sym.IsTypeDef(), "Invalid definition of type")
	structDef := sym.(*TypeDef)
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

		fieldType := c.LookupSym(field.Type)

		if fieldType == nil || !fieldType.IsTypeDef() {
			start, end := field.Node.Type.Pos()
			c.errorf(start, end, "type %s is not defined", field.Type)
		}

	}
}

func (c *Checker) GetVariable(name string) Symbol {
	if sym, ok := c.Scope.Symbols[name]; ok {
		return sym
	}
	return nil
}

func checkTypeCompability(typ1, exprType string) bool {

	if exprType == "float" {
		return typ1 == "f32" || typ1 == "f64" || typ1 == "f16"
	}

	if exprType == "int" {
		return typ1 == "i8" || typ1 == "i16" || typ1 == "i32" || typ1 == "i64"
	}

	return typ1 == exprType
}

func getBitSize(typName string) (int, error) {
	if typName == "i32" || typName == "f32" {
		return 32, nil
	} else if typName == "i16" || typName == "f16" {
		return 16, nil
	} else if typName == "i8" {
		return 8, nil
	} else if typName == "f64" || typName == "i64" {
		return 64, nil
	}

	return 0, fmt.Errorf("not a number")
}

func (c *Checker) checkVarDef(stmt *VarDefStmt) {
	var isErr bool
	variable := c.GetVariable(stmt.Name.Name)
	if variable != nil {
		start, end := stmt.Name.Pos()
		c.errorf(start, end, "redeclared symbol %s", stmt.Name.Name)
		isErr = true
	}

	typName := c.getNameOfType(stmt.Type)

	typSym := c.LookupSym(typName)
	bitSize, err := getBitSize(typName)
	if typSym == nil || !typSym.IsTypeDef() {
		start, end := stmt.Type.Pos()
		c.errorf(start, end, "type %s is not defined", typName)
		isErr = true
	}

	if stmt.Init != nil {
		if err == nil {
			c.pushExpectedBitSize(bitSize)
		}
		r, err := c.checkExpr(stmt.Init)
		if err != nil {
			return
		}

		if err == nil {
			c.popExpectedBitSize()
		}
		if !checkTypeCompability(typName, r) && typSym != nil {
			start, end := stmt.Init.Pos()
			c.errorf(start, end, "Expected type %s but got %s", typName, r)
			isErr = true
		}
	}

	if !isErr {
		c.DefineVar(stmt.Name.Name, typName)
	}
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

	bitSize, err := getBitSize(funcdef.Type)

	typ := c.LookupSym(funcdef.Type).(*TypeDef)

	if typ == nil {
		start, end := funcdef.DefNode.Type.Pos()
		c.errorf(start, end, "Type %s is not defined", funcdef.Type)
	} else {
		c.expectedFuncType = typ.Name
		if err == nil {
			c.pushExpectedBitSize(bitSize)
		}
	}

	for _, param := range funcdef.Params {
		if c.GetVariable(param.Name) == nil {
			c.DefineVar(param.Name, param.Type)
		} else {
			start, end := param.Param.Pos()
			c.errorf(start, end, "re defined param %s", param.Name)
		}

		paramType := c.LookupSym(param.Type)
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
		exprType, err := c.checkExpr(stmt.Value)
		if err != nil {
			return
		}
		if !checkTypeCompability(c.expectedFuncType, exprType) {
			start, end := stmt.Value.Pos()
			c.errorf(start, end, "expected %s, got %s", c.expectedFuncType, exprType)
		}

	}
}

func isNumber(t string) bool {
	return t == "i8" || t == "i16" || t == "i32" || t == "i64" || t == "f16" || t == "f32" || t == "f64"
}

// checkExpr checks expr and returns its type and error if it exists
func (c *Checker) checkExpr(expr Expr) (string, error) {
	switch expr := expr.(type) {
	case *NumberExpr:
		size := c.topExpectedBitSize()
		expr.BitSize = size
		return "i" + strconv.Itoa(size), nil
	case *FloatExpr:
		size := c.topExpectedBitSize()
		expr.BitSize = size
		return "f" + strconv.Itoa(size), nil
	case *StringExpr:
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
				if !checkTypeCompability(firstType, elemType) {
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
		return lookupTyp.(*TypeDef).Name, nil
	case *BinaryExpr:
		c.dupExpectedBitSize()
		left, err := c.checkExpr(expr.Left)
		if err != nil {
			return left, err
		}

		right, err := c.checkExpr(expr.Right)
		if err != nil {
			return right, err
		}

		c.popExpectedBitSize()
		// TODO compability cases
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

		typDef := c.LookupSym(left)
		if typDef == nil || !typDef.IsTypeDef() {
			start, end := expr.Selector.Pos()
			c.errorf(start, end, "provided type %s is not defined", left)
			return unresolvedType, fmt.Errorf("unknown type %s", left)
		}

		c.enterTypeScope(left)
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
		typedef := c.LookupSym(sym.TypeName())
		return typedef.TypeName(), nil
	case *NilExpr:
		return "pointer<nil>", nil
	case *IndexExpr:
		panic("not implemented indexExpr")
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

			var r string
			_, err = fmt.Sscanf(right, pointerScheme, &r)
			AssertErr(err)
			return r, nil
		case Band:
			panic("unhandled band")
		case Sizeof:
			panic("unhandled sizeof")
		default:
			panic("unreachable")
		}
	case *CallExpr:
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
		if !funcDef.TypeCast {
			for i, _ := range expr.Args {

				var err3 error
				if i < len(funcDef.Params) {
					var size int
					size, err3 = getBitSize(funcDef.Params[i].Type)
					if err3 == nil {
						c.pushExpectedBitSize(size)
					}
				}

				t, err2 := c.checkExpr(expr.Args[i])
				if err2 != nil {
					continue
				}

				if i >= len(funcDef.Params) {
					break
				}

				arg := funcDef.Params[i]

				if err == nil {
					err = err2
				}

				if err3 == nil {
					c.popExpectedBitSize()
				}

				if !checkTypeCompability(arg.Type, t) {
					start, end := expr.Args[i].Pos()
					c.errorf(start, end, "expected %s, got %s", arg.Type, t)
					err = fmt.Errorf("expected %s, got %s", arg.Type, t)
				}
			}
		} else {
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
					}

				}

			}

		}
		return l, err
	default:
		panic("unreachable")
	}

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
