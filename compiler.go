package main

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"os"
)

type Scope struct {
	variables map[string]value.Value
	Parent    *Scope
	Children  []*Scope
	// TODO Shorthands map[string]value.Value // Like this.a -> just a if it is not defined in the scope
}

func (s *Scope) Lookup(name string) value.Value {
	if s.variables[name] != nil {
		return s.variables[name]
	}

	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}

	panic("No variable found: " + name)
}

func (s *Scope) Define(name string, v value.Value) {
	if s.variables[name] != nil {
		panic("Duplicate variable: " + name)
	}

	s.variables[name] = v
}

const entryBlock string = "entry"

type fieldIndexKey struct {
	field, typ string
}

type Compiler struct {
	module        *ir.Module
	currentScope  *Scope
	currentBlock  *ir.Block
	breakBlock    *ir.Block
	continueBlock *ir.Block
	currentFunc   *ir.Func
	currentType   types.Type
	currentAlloc  *ir.InstAlloca

	funcs         map[string]*ir.Func
	types         map[string]types.Type
	typesIndexMap map[fieldIndexKey]int
	globals       map[string]value.Value
	pkg           *Package
}

func (c *Compiler) Define(name string, v value.Value) {
	c.currentScope.Define(name, v)
}

func (c *Compiler) Lookup(name string) value.Value {
	return c.currentScope.Lookup(name)
}

func (c *Compiler) compile() {
	Assert(c.pkg != nil, "pkg is nil")
	Assert(c.module != nil, "module is nil")
	for _, stmt := range c.pkg.Stmts {
		c.compileStmt(stmt)
	}

}

func (c *Compiler) initBuiltinTypes() {
	// we have already built-in types in llvm, so we don't have to re-define them
	// just add them to the dictionary so that make them accessible in the compiler
	if c.types == nil {
		c.types = make(map[string]types.Type)
		c.typesIndexMap = make(map[fieldIndexKey]int)
	}

	c.types["void"] = types.Void
	c.types["int"] = types.I32
	c.types["bool"] = types.I1

	stringStruct := types.NewStruct(types.NewPointer(types.I8), types.I64)
	def := c.module.NewTypeDef("string", stringStruct)
	c.types["string"] = def
}

func (c *Compiler) initBuiltinFuncs() {
	module := c.module

	if c.funcs == nil {
		c.funcs = make(map[string]*ir.Func)
	}

	runtimePrintf := module.NewFunc("r_runtime_printf", types.I32,
		ir.NewParam("", types.NewPointer(types.I8)))
	runtimePrintf.Sig.Variadic = true
	runtimePrintf.Linkage = enum.LinkageExternal

	c.funcs["r_runtime_printf"] = runtimePrintf

	runtimeScanf := module.NewFunc("r_runtime_scanf", types.I32, ir.NewParam("", types.NewPointer(types.I8)))
	runtimeScanf.Sig.Variadic = true
	runtimeScanf.Linkage = enum.LinkageExternal
	c.funcs["r_runtime_scanf"] = runtimeScanf

	runtimeExit := module.NewFunc("r_runtime_exit", types.Void, ir.NewParam("", types.I32))
	runtimeExit.Linkage = enum.LinkageExternal
	c.funcs["r_runtime_exit"] = runtimeExit

	printfInternal := module.NewFunc("printf_internal", types.I32, ir.NewParam("", c.types["string"]))
	printfInternal.Sig.Variadic = true
	printfInternal.Linkage = enum.LinkageExternal
	c.funcs["printf_internal"] = printfInternal

	makeString := module.NewFunc("make_string", c.types["string"], ir.NewParam("", types.NewPointer(types.I8)))
	makeString.Linkage = enum.LinkageExternal
	c.funcs["make_string"] = makeString

	scanfInternal := module.NewFunc("scanf_internal", types.I32, ir.NewParam("", c.types["string"]))
	scanfInternal.Sig.Variadic = true
	scanfInternal.Linkage = enum.LinkageExternal
	c.funcs["scanf_internal"] = scanfInternal

	strlenInternal := module.NewFunc("strlen_internal", types.I64, ir.NewParam("", c.types["string"]))
	strlenInternal.Linkage = enum.LinkageExternal
	c.funcs["strlen_internal"] = strlenInternal

	mallocInternal := module.NewFunc("malloc_internal", types.NewPointer(types.I8), ir.NewParam("size", types.I64))
	mallocInternal.Linkage = enum.LinkageExternal
	c.funcs["malloc_internal"] = mallocInternal

}

func (c *Compiler) resolveType(expr Expr) types.Type {
	switch expr := expr.(type) {
	case *IdentExpr:
		if typ, ok := c.types[expr.Name]; ok {
			return typ
		}

		panic(fmt.Sprintf("no provided type %s", expr.Name))
	case *RefTypeExpr:
		typ := c.resolveType(expr.Expr)
		return types.NewPointer(typ)
	case *ArrayTypeExpr:
		typ := c.resolveType(expr.Type)
		return types.NewArray(uint64(expr.Size.(*NumberExpr).Value), typ)
	}

	panic("unreachable")
}

func (c *Compiler) defineType(expr Expr) types.Type {
	Assert(c.types != nil, "un-initialized type map")

	panic("TODO")
}

func (c *Compiler) resetTempFields() {
	c.currentAlloc = nil
	c.currentType = nil
}

func (c *Compiler) compileVarDef(stmt *VarDefStmt) {
	Assert(c.currentScope != nil, "un-initialized scope")
	Assert(c.currentBlock != nil, "currentBlock cannot be nil")
	typ := c.resolveType(stmt.Type)
	c.currentType = typ
	ptr := c.currentBlock.NewAlloca(typ)
	c.currentAlloc = ptr
	defer c.resetTempFields()
	c.Define(stmt.Name.Name, ptr)

	if stmt.Init != nil {
		right := c.compileExpr(stmt.Init)
		switch right.Type().(type) {
		case *types.PointerType:
			cast := c.currentBlock.NewBitCast(right, typ)
			c.currentBlock.NewStore(cast, ptr)
		default:
			c.currentBlock.NewStore(right, ptr)
		}
	}
}

func (c *Compiler) compileFunc(def *FuncDefStmt, isMethod bool, thisType types.Type) *ir.Func {
	Assert(def.Name != nil, "type checker has to handle this")
	Assert(def.Name.Name != "", "name can't be empty")
	name := def.Name.Name
	c.currentScope = &Scope{
		variables: make(map[string]value.Value),
		Parent:    c.currentScope,
		Children:  make([]*Scope, 0),
	}
	typ := c.resolveType(def.Type)

	params := make([]*ir.Param, 0)

	if isMethod {
		param := ir.NewParam("this", thisType)
		params = append(params, param)
		c.Define("this", param)
	}

	for _, def := range def.Params {
		Assert(def.Name.Name != "this", "param name cannot be this")
		typ2 := c.resolveType(def.Type)
		param := ir.NewParam(def.Name.Name, typ2)
		params = append(params, param)

		// TODO merge with compileVarDef function
		c.Define(def.Name.Name, param)
	}
	f := c.module.NewFunc(name, typ, params...)
	c.funcs[name] = f
	c.currentFunc = f
	if !def.IsExternal {
		block := f.NewBlock(entryBlock)
		c.currentBlock = block

		for _, innerStmt := range def.Body.Stmts {
			c.compileStmt(innerStmt)
		}
	} else {
		Assert(!isMethod, "Methods cannot be external")
		f.Linkage = enum.LinkageExternal
	}

	return f
}

func (c *Compiler) compileStructDef(def *StructDefStmt) {
	str := types.NewStruct()
	for i, field := range def.Fields {
		typ := c.resolveType(field.Type)
		str.Fields = append(str.Fields, typ)
		c.typesIndexMap[fieldIndexKey{field.Name.Name, def.Name.Name}] = i
	}
	td := c.module.NewTypeDef(def.Name.Name, str)
	c.types[def.Name.Name] = td
}

func (c *Compiler) compileStmt(stmt Stmt) {
	switch innerStmt := stmt.(type) {
	case *ReturnStmt:
		if innerStmt.Value == nil {
			c.currentBlock.NewRet(nil)
		} else {
			c.currentBlock.NewRet(c.compileExpr(innerStmt.Value))
		}
	case *BreakStmt:
		c.currentBlock.NewBr(c.breakBlock)
	case *ContinueStmt:
		c.currentBlock.NewBr(c.continueBlock)
	case *ExprStmt:
		c.compileExpr(innerStmt.Expr)
	case *IfStmt:
		c.compileIfStmt(innerStmt)
	case *FuncDefStmt:
		c.compileFunc(innerStmt, false, nil)
	case *VarDefStmt:
		c.compileVarDef(innerStmt)
	case *ForStmt:
		c.compileForStmt(innerStmt)
	case *StructDefStmt:
		c.compileStructDef(innerStmt)
	case *BlockStmt:
		for _, innerStmt := range innerStmt.Stmts {
			c.compileStmt(innerStmt)
		}
	case *ImplStmt:
		thisType := c.resolveType(innerStmt.Type)

		for _, s := range innerStmt.Stmts {
			switch implStmt := s.(type) {
			case *FuncDefStmt:
				implStmt.ImplOf = innerStmt // make sure the correct impl is used
				c.compileFunc(implStmt, true, thisType)
			}
		}

	default:
		panic("unreachable compileStmt on compiler")

	}
}

func (c *Compiler) compileIfStmt(stmt *IfStmt) {
	cond := c.compileExpr(stmt.Cond)
	thenBlock := c.currentFunc.NewBlock("")
	continueBlock := c.currentFunc.NewBlock("")
	var elseBlock *ir.Block = continueBlock
	if stmt.Else != nil {
		elseBlock = c.currentFunc.NewBlock("")
	}

	c.currentBlock.NewCondBr(cond, thenBlock, elseBlock)
	c.currentBlock = thenBlock
	c.compileStmt(stmt.Then)

	if thenBlock.Term == nil {
		thenBlock.NewBr(continueBlock)
	}

	if stmt.Else != nil {
		c.currentBlock = elseBlock
		c.compileStmt(stmt.Else)
	}

	if c.currentBlock.Term == nil {
		c.currentBlock.NewBr(continueBlock)
	}

	c.currentBlock = continueBlock
}

/*

	initBlock -> condBlock ->  forBlock -> stepBlock -> condBlock
							-> funcBlock

	initBlock -> condBlock ->  forBlock -> ifBlock -> continueBlock -> stepBlock -> condBlock
							-> funcBlock
*/

func (c *Compiler) compileForStmt(stmt *ForStmt) {
	funcBlock := c.currentFunc.NewBlock("")
	initBlock := c.currentFunc.NewBlock("")
	forBlock := c.currentFunc.NewBlock("")
	condBlock := c.currentFunc.NewBlock("")
	stepBlock := c.currentFunc.NewBlock("")

	c.continueBlock = condBlock
	c.breakBlock = funcBlock

	defer func() {
		c.continueBlock = nil
		c.breakBlock = nil
	}()

	c.currentBlock.NewBr(initBlock)
	c.currentBlock = initBlock
	if stmt.Init != nil {
		c.compileStmt(stmt.Init)
	}
	initBlock.NewBr(condBlock)

	c.currentBlock = condBlock
	condBlock.NewCondBr(c.compileExpr(stmt.Cond), forBlock, funcBlock)
	if stmt.Step != nil {
		c.currentBlock = stepBlock
		c.continueBlock = stepBlock
		c.compileExpr(stmt.Step)
		stepBlock.NewBr(condBlock)
	} else {
		stepBlock.NewBr(condBlock)
	}

	c.currentBlock = forBlock

	c.compileStmt(stmt.Body)
	if c.currentBlock != forBlock && c.currentBlock.Term == nil {
		c.currentBlock.NewBr(c.continueBlock)
	}
	if forBlock.Term == nil {
		c.currentBlock.NewBr(c.continueBlock)
	}

	c.currentBlock = funcBlock
}

func (c *Compiler) getAlloca(name string) value.Value {
	v := c.Lookup(name)
	switch v := v.(type) {
	case *ir.InstAlloca:
		return v
	case *ir.Param:
		return v
	}
	return v
}

func (c *Compiler) call(expr *CallExpr) value.Value {
	b := c.currentBlock

	fun := c.getFunc(expr.Left.(*IdentExpr).Name)

	args := make([]value.Value, 0)
	for _, e := range expr.Args {
		arg := c.compileExpr(e)
		args = append(args, arg)
	}

	return b.NewCall(fun, args...)
}

func (c *Compiler) resolveFieldIndex(name, typ string) (int, error) {
	if index, ok := c.typesIndexMap[fieldIndexKey{name, typ}]; ok {
		return index, nil
	}

	return -1, fmt.Errorf("unknown field index: %v for %s", name, typ)
}

// sizeOfCompilationTime returns the size of
func sizeof(p types.Type) (int64 uint64) {
	switch p := p.(type) {
	case *types.IntType:
		return p.BitSize / 8
	default:
		panic("TODO unreachable sizeof")
	}
}

func (c *Compiler) newArray(expr *ArrayLitExpr, size uint64, typ types.Type) *ir.InstLoad {
	var alloc = c.currentAlloc
	if alloc == nil {
		alloc = c.currentBlock.NewAlloca(typ)
	}
	elems := expr.Elems
	for i, elem := range elems {
		e := c.compileExpr(elem)
		zero := constant.NewInt(types.I64, 0)
		index := constant.NewInt(types.I64, int64(i))
		gep := c.currentBlock.NewGetElementPtr(typ, alloc, zero, index)
		gep.InBounds = true
		c.currentBlock.NewStore(e, gep)
	}

	return c.currentBlock.NewLoad(typ, alloc)
}

func (c *Compiler) newObject(expr *ObjectLitExpr, alloc *ir.InstAlloca) *ir.InstLoad {
	typ := c.resolveType(expr.Type)
	c.resetTempFields()

	for _, kv := range expr.KeyValue {
		key, ok := kv.Key.(*IdentExpr)
		Assert(ok, "Key must be a identifier")
		fieldIndex, err := c.resolveFieldIndex(key.Name, typ.Name())
		AssertErr(err)
		v := c.compileExpr(kv.Value)
		gep := c.currentBlock.NewGetElementPtr(typ, alloc, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, int64(fieldIndex)))
		c.currentBlock.NewStore(v, gep)
	}
	return c.currentBlock.NewLoad(typ, alloc)
}

func (c *Compiler) getIndexedType(load *ir.InstLoad, idx int) types.Type {
	switch l := load.ElemType.(type) {
	case *types.StructType:
		return l.Fields[idx]
	default:
		panic("Unreachable indexed type")
	}
}

func (c *Compiler) compileExpr(expr Expr) value.Value {
	switch expr := expr.(type) {
	case *NilExpr:
		null := constant.NewNull(types.NewPointer(types.I8))
		return null
	case *ArrayLitExpr:
		return c.newArray(expr, uint64(len(expr.Elems)), types.NewArray(2, c.resolveType(&IdentExpr{Name: "User"})))
	case *BinaryExpr:
		left := c.compileExpr(expr.Left)
		right := c.compileExpr(expr.Right)
		var v value.Value = left
		switch expr.Left.(type) {
		case *UnaryExpr:
			v = c.currentBlock.NewLoad(types.I32, v)
		}

		return generateOperation(c.currentBlock, v, right, expr.Op)
	case *NumberExpr:
		return getI32Constant(expr.Value)
	case *StringExpr:
		v := expr.Unquoted
		v2 := constant.NewCharArrayFromString(v)
		strPtr := c.module.NewGlobalDef("", v2)
		strPtr.Immutable = true
		strPtr.Align = ir.Align(1)
		strPtr.UnnamedAddr = enum.UnnamedAddrUnnamedAddr
		var zero = constant.NewInt(types.I8, 0)
		gep := constant.NewGetElementPtr(v2.Typ, strPtr, zero, zero)
		return c.currentBlock.NewCall(c.getFunc("make_string"), gep)
	case *ObjectLitExpr:
		alloc := c.currentAlloc
		if alloc == nil {
			alloc = c.currentBlock.NewAlloca(c.resolveType(expr.Type))
		}
		return c.newObject(expr, alloc)
	case *CallExpr:
		return c.call(expr)
	case *IdentExpr:
		v := c.getAlloca(expr.Name)
		switch v := v.(type) {
		case *ir.InstAlloca:
			return c.currentBlock.NewLoad(v.ElemType, v)
		case *ir.Param:
			// TODO this is workaround
			alloca := c.currentBlock.NewAlloca(v.Typ)
			c.currentBlock.NewStore(v, alloca)
			return c.currentBlock.NewLoad(alloca.ElemType, alloca)
		default:
			panic("unreachable")
		}
	case *ParenExpr:
		return c.compileExpr(expr.Expr)
	case *BoolExpr:
		return constant.NewBool(expr.Value)
	case *UnaryExpr:
		switch expr.Op {
		case Sub:
			right := c.compileExpr(expr.Right)
			return c.currentBlock.NewMul(constant.NewInt(types.I32, -1), right)
		case Band:
			val := c.compileExpr(expr.Right)
			return val.(*ir.InstLoad).Src
		case Sizeof:
			// TODO constant expressions must, we handle as statically to save day for now
			resolvedType := c.resolveType(expr.Right)
			var size = sizeof(resolvedType)
			return constant.NewInt(types.I64, int64(size))
		case Mul:
			addr := c.compileExpr(expr.Right)
			load, ok := addr.(*ir.InstLoad)
			Assert(ok, "*(ptr) operand inst load problem")
			typ, ok := load.ElemType.(*types.PointerType)
			Assert(ok, "* operator must have combined with pointer")
			elemType := typ.ElemType
			return c.currentBlock.NewLoad(elemType, addr)
		default:
			panic("Unreachable unary expression op = " + expr.Op.String())
		}
	case *AssignExpr:
		left := c.compileExpr(expr.Left)
		right := c.compileExpr(expr.Right)
		load, ok := left.(*ir.InstLoad)
		Assert(ok, "assign_expr left operand is not loaded")
		Assert(load.Src != nil, "assign_expr load.Src is nil")
		c.currentBlock.NewStore(right, load.Src)
		return left
	case *IndexExpr:
		left := c.compileExpr(expr.Left)
		index := c.compileExpr(expr.Index)
		zero := constant.NewInt(types.I64, 0)
		leftInstLoad, ok := left.(*ir.InstLoad)
		Assert(ok, "IndexExpr left operand is invalid.")
		gep := c.currentBlock.NewGetElementPtr(leftInstLoad.ElemType, leftInstLoad.Src, zero, index)
		arrayType, ok := leftInstLoad.ElemType.(*types.ArrayType)
		Assert(ok, "Index operation cannot be used with non-array type")
		return c.currentBlock.NewLoad(arrayType.ElemType, gep)
	case *SelectorExpr:
		sel := c.compileExpr(expr.Selector)
		selLoad, ok := sel.(*ir.InstLoad)
		Assert(ok, "Selector invalid")
		zero := constant.NewInt(types.I32, 0)
		fieldIndex, err := c.resolveFieldIndex(expr.Ident.(*IdentExpr).Name, sel.Type().Name())
		AssertErr(err)
		idx := constant.NewInt(types.I32, int64(fieldIndex))
		gep := c.currentBlock.NewGetElementPtr(selLoad.ElemType, selLoad.Src, zero, idx)
		return c.currentBlock.NewLoad(c.getIndexedType(selLoad, fieldIndex), gep)
	default:
		panic("unreachable")
	}
}

type Operation int

const (
	Add Operation = iota
	Sub
	Mul
	Div
	Mod
	Eq
	Neq
	Lt
	Lte
	Gt
	Gte
	Not
	And
	Band
	Or
	Sizeof
)

func (o Operation) String() string {
	switch o {
	case Add:
		return "+"
	case Sub:
		return "-"
	case Mul:
		return "*"
	case Div:
		return "/"
	case Mod:
		return "%"
	case Eq:
		return "=="
	case Neq:
		return "!="
	case Lt:
		return "<"
	case Lte:
		return "<="
	case Gt:
		return ">"
	case Gte:
		return ">="
	case Not:
		return "!"
	case Band:
		return "&"
	case Sizeof:
		return "sizeof"
	default:
		panic("unreachable")
	}
}

type TokenKind int

func debug(s string, args ...interface{}) {
	if os.Getenv("DEBUG") == "" {
		return
	}
	fmt.Print("[DEBUG] ")
	fmt.Printf(s, args...)
	fmt.Println()
}

func getI32Constant(i int64) constant.Constant {
	return constant.NewInt(types.I32, i)
}

func generateOperation(block *ir.Block, value1, value2 value.Value, op Operation) value.Value {
	debug("operation value1 = %v, value2 = %v, OP = %s", value1, value2, op)
	switch op {
	case Add:
		return block.NewAdd(value1, value2)
	case Sub:
		return block.NewSub(value1, value2)
	case Mul:
		return block.NewMul(value1, value2)
	case Div:
		return block.NewUDiv(value1, value2)
	case Mod:
		return block.NewURem(value1, value2)
	case Eq:
		return block.NewICmp(enum.IPredEQ, value1, value2)
	case Neq:
		return block.NewICmp(enum.IPredNE, value1, value2)
	case Gt:
		return block.NewICmp(enum.IPredSGT, value1, value2)
	case Gte:
		return block.NewICmp(enum.IPredSGE, value1, value2)
	case Lt:
		return block.NewICmp(enum.IPredSLT, value1, value2)
	case Lte:
		return block.NewICmp(enum.IPredSLE, value1, value2)
	case And:
		return block.NewAnd(value1, value2)
	case Or:
		return block.NewOr(value1, value2)
	}

	panic("unreachable")

}

func (c *Compiler) init() {
	Assert(c.module != nil, "module not initialized")

	c.currentScope = &Scope{
		variables: make(map[string]value.Value),
		Parent:    nil,
		Children:  make([]*Scope, 0),
	}

	c.initBuiltinTypes()
	c.initBuiltinFuncs()
}

func (c *Compiler) getFunc(name string) *ir.Func {
	if f, ok := c.funcs[name]; ok {
		return f
	}
	return nil
}
