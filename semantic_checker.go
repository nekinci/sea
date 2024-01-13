package main

import (
	"fmt"
	"strings"
)

const arrayScheme = "array<%s, %d>"
const pointerScheme = "pointer<%s>"
const unresolvedType = "<unresolved>"

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

type VarDef struct {
	DefNode *VarDefStmt
	Name    string
	Type    string
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
}

type Scope struct {
	TypeDefs   map[string]*TypeDef
	VarDefs    map[string]*VarDef
	FuncDefs   map[string]*FuncDef
	Parent     *Scope
	Children   []*Scope
	funcScopes map[string]*Scope
}

func (c *Checker) EnterScope() *Scope {
	newScope := &Scope{
		TypeDefs:   make(map[string]*TypeDef),
		VarDefs:    make(map[string]*VarDef),
		FuncDefs:   make(map[string]*FuncDef),
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

func (scope *Scope) LookupFuncDef(name string) *FuncDef {
	if f := scope.FuncDefs[name]; f != nil {
		return f
	}

	if scope.Parent != nil {
		return scope.Parent.LookupFuncDef(name)
	}

	return nil
}

func (scope *Scope) GetVariable(name string) *VarDef {
	return scope.VarDefs[name]
}

func (scope *Scope) DefineVar(name string, typ string) *VarDef {
	scope.VarDefs[name] = &VarDef{
		DefNode: nil,
		Name:    name,
		Type:    typ,
	}
	return scope.VarDefs[name]
}

func (scope *Scope) LookupVariable(name string) *VarDef {
	if v := scope.GetVariable(name); v != nil {
		return v
	}

	if scope.Parent != nil {
		return scope.Parent.LookupVariable(name)
	}

	return nil
}

func (scope *Scope) LookupTypeDef(name string) *TypeDef {
	if def, ok := scope.TypeDefs[name]; ok {
		return def
	}

	if scope.Parent != nil {
		return scope.Parent.LookupTypeDef(name)
	}

	return nil
}

type Checker struct {
	*Scope
	Package     *Package
	Errors      []Error
	currentFile string
	typeScopes  map[string]*Scope

	// Context
	expectedFuncType string
	inCallExpr       bool
	currentFuncSym   *FuncDef
	currentTypeSym   *TypeDef
	currentVarSym    *VarDef
	//
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

func (c *Checker) initGlobalScope() {
	c.TypeDefs["int"] = &TypeDef{
		DefNode:   nil,
		Name:      "int",
		Fields:    nil,
		Methods:   nil,
		Completed: true,
	}
	c.TypeDefs["string"] = &TypeDef{
		DefNode:   nil,
		Name:      "string",
		Fields:    nil,
		Methods:   nil,
		Completed: true,
	}

	c.FuncDefs["printf_internal"] = &FuncDef{
		DefNode: nil,
		Name:    "printf_internal",
		Type:    "int",
		Params: []Param{{
			Param: nil,
			Name:  "fmt",
			Type:  "string",
		}},
		External:  true,
		Variadic:  true,
		Completed: false,
	}

	c.FuncDefs["scanf_internal"] = &FuncDef{
		DefNode: nil,
		Name:    "scanf_internal",
		Type:    "int",
		Params: []Param{{
			Param: nil,
			Name:  "fmt",
			Type:  "string",
		}},
		External:  true,
		Variadic:  true,
		Completed: false,
	}
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
	structDef := c.LookupTypeDef(stmt.Name.Name)
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
		}

		if c.LookupTypeDef(field.Type) == nil {
			start, end := field.Node.Type.Pos()
			c.errorf(start, end, "Provided type %s is not implemented", field.Type)
		}

	}
}

func (c *Checker) checkVarDef(stmt *VarDefStmt) {
	var isErr bool
	variable := c.GetVariable(stmt.Name.Name)
	if variable != nil {
		start, end := stmt.Name.Pos()
		c.errorf(start, end, "re defined variable %s", stmt.Name.Name)
		isErr = true
	}

	typName := c.getNameOfType(stmt.Type)

	typDef := c.LookupTypeDef(typName)
	if typDef == nil {
		start, end := stmt.Type.Pos()
		c.errorf(start, end, "Type %s is not defined", typName)
		isErr = true
	}

	if stmt.Init != nil {
		r, err := c.checkExpr(stmt.Init)
		if err != nil {
			return
		}

		if r != typName && typDef != nil {
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
	funcdef := c.LookupFuncDef(stmt.Name.Name)
	defer func() {
		funcdef.Completed = true
	}()

	if funcdef.Completed {
		return
	}

	typ := c.LookupTypeDef(funcdef.Type)

	if typ == nil {
		start, end := funcdef.DefNode.Type.Pos()
		c.errorf(start, end, "Type %s is not defined", funcdef.Type)
	} else {
		c.expectedFuncType = typ.Name
	}

	for _, param := range funcdef.Params {
		if c.GetVariable(param.Name) == nil {
			c.DefineVar(param.Name, param.Type)
		} else {
			start, end := param.Param.Pos()
			c.errorf(start, end, "re defined param %s", param.Name)
		}

		if c.LookupTypeDef(param.Type) == nil {
			start, end := param.Param.Type.Pos()
			c.errorf(start, end, "Provided type %s is not defined", param.Type)
		}
	}

	if funcdef.External {
		return
	}

	c.checkBlockStmt(stmt.Body)
}

func (c *Checker) checkStmt(stmt Stmt) {
	switch stmt := stmt.(type) {
	case *BlockStmt:
		c.checkBlockStmt(stmt)
	case *VarDefStmt:
		c.checkVarDef(stmt)
	case *ExprStmt:
		c.checkExprStmt(stmt)
	case *ReturnStmt:
		exprType, err := c.checkExpr(stmt.Value)
		if err != nil {
			return
		}
		if c.expectedFuncType != exprType {
			start, end := stmt.Value.Pos()
			c.errorf(start, end, "expected %s, got %s", c.expectedFuncType, exprType)
		}

	}
}

// checkExpr checks expr and returns its type and error if it exists
func (c *Checker) checkExpr(expr Expr) (string, error) {
	switch expr := expr.(type) {
	case *NumberExpr:
		return "int", nil
	case *StringExpr:
		return "string", nil
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
				if firstType != elemType {
					start, end := elem.Pos()
					c.errorf(start, end, "expected %s, got %s", firstType, elemType)
				}
			}
		}
		return fmt.Sprintf(arrayScheme, firstType, len(expr.Elems)), nil
	case *ObjectLitExpr:
		typ := c.getNameOfType(expr.Type)
		lookupTyp := c.LookupTypeDef(typ)
		if lookupTyp == nil {
			start, end := expr.Type.Pos()
			c.errorf(start, end, "provided type %s is not defined", typ)
			return unresolvedType, fmt.Errorf("unknown type %s", typ)
		}
		return lookupTyp.Name, nil
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
		if left != right {
			start, end := expr.Pos()
			c.errorf(start, end, "types are not compatible %s, %s", left, right)
			return unresolvedType, fmt.Errorf("types are not compatible %s, %s", left, right)
		}

		return left, nil
	case *SelectorExpr:
		// u.age2()
		left, err := c.checkExpr(expr.Selector)
		if err != nil {
			return unresolvedType, err
		}

		right, err := c.checkExpr(expr.Ident)
		if err != nil {
			return unresolvedType, err
		}

		_, _ = left, right

		panic("unreachable")

	case *IdentExpr:

		if c.inCallExpr {
			def := c.LookupFuncDef(expr.Name)
			c.currentFuncSym = def
			if def == nil {
				start, end := expr.Pos()
				c.errorf(start, end, "function %s not found", expr.Name)
				return unresolvedType, fmt.Errorf("function %s not found", expr.Name)
			}

			return c.LookupTypeDef(def.Type).Name, nil
		}

		variable := c.LookupVariable(expr.Name)
		if variable == nil {
			start, end := expr.Pos()
			c.errorf(start, end, "variable %s not found", expr.Name)
			return unresolvedType, fmt.Errorf("variable %s not found", expr.Name)
		}

		typedef := c.LookupTypeDef(variable.Type)
		return typedef.Name, nil
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

			if right != "int" {
				start, end := expr.Pos()
				c.errorf(start, end, "right operand must be int")
				return unresolvedType, fmt.Errorf("right operand must be int")
			}

			return "int", nil
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
		c.inCallExpr = true
		l, err := c.checkExpr(expr.Left)
		defer func() {
			c.currentFuncSym = nil
		}()
		c.inCallExpr = false
		if err != nil {
			return unresolvedType, err
		}

		funcDef := c.currentFuncSym

		if !funcDef.Variadic && len(funcDef.Params) != len(expr.Args) {
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

		for i, _ := range expr.Args {
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

			if t != arg.Type {
				start, end := expr.Args[i].Pos()
				c.errorf(start, end, "expected %s, got %s", arg.Type, t)
				err = fmt.Errorf("expected %s, got %s", arg.Type, t)
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
		case *StructDefStmt:
			c.checkStructDef(stmt)
		case *VarDefStmt:
			c.checkVarDef(stmt)
		case *FuncDefStmt:
			c.checkFuncDef(stmt)
		}
	}
}

func (c *Checker) collectSignatures() {
	for _, stmt := range c.Package.Stmts {
		switch stmt := stmt.(type) {
		case *StructDefStmt:

			if _, ok := c.TypeDefs[stmt.Name.Name]; ok {
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

			c.TypeDefs[stmt.Name.Name] = typeDef
		case *VarDefStmt:
			if _, ok := c.VarDefs[stmt.Name.Name]; ok {
				panic("Not implemented")
			}
			c.VarDefs[stmt.Name.Name] = &VarDef{
				DefNode: stmt,
				Name:    stmt.Name.Name,
				Type:    c.getNameOfType(stmt.Type),
			}
		case *FuncDefStmt:
			if _, ok := c.FuncDefs[stmt.Name.Name]; ok {
				panic("Not implemented")
			}

			funcDef := &FuncDef{
				DefNode:  stmt,
				Name:     stmt.Name.Name,
				Type:     c.getNameOfType(stmt.Type),
				External: stmt.IsExternal,
			}

			params := make([]Param, 0)
			for _, param := range stmt.Params {
				params = append(params, Param{
					Name:  param.Name.Name,
					Type:  c.getNameOfType(param.Type),
					Param: param,
				})
			}
			funcDef.Params = params
			c.FuncDefs[stmt.Name.Name] = funcDef
		}
	}
}
