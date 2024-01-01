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

const entryBlock string = "entry"

type Compiler struct {
	module        *ir.Module
	variables     map[string]value.Value
	currentBlock  *ir.Block
	breakBlock    *ir.Block
	continueBlock *ir.Block

	currentFunc *ir.Func
	funcs       map[string]*ir.Func
	types       map[string]types.Type
	globals     map[string]value.Value
	pkg         *Package
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
	}

	panic("TODO")
}

func (c *Compiler) defineType(expr Expr) types.Type {
	Assert(c.types != nil, "un-initialized type map")

	panic("TODO")
}

func (c *Compiler) compileVarDef(stmt *VarDefStmt) {
	Assert(c.variables != nil, "un-initialized variable map")
	Assert(c.currentBlock != nil, "currentBlock cannot be nil")
	block := c.currentBlock

	typ := c.resolveType(stmt.Type)
	if stmt.IsPtr {
		typ = types.NewPointer(typ)
	}
	ptr := block.NewAlloca(typ)
	c.variables[stmt.Name.Name] = ptr

	if stmt.Init != nil {
		if _, ok := stmt.Init.(*NilExpr); ok {
			block.NewStore(constant.NewNull(typ.(*types.PointerType)), ptr)
		} else {
			if stmt.IsPtr {
				e := c.compileExpr(stmt.Init)
				cast := c.currentBlock.NewBitCast(e, typ)
				block.NewStore(cast, ptr)
			} else {
				block.NewStore(c.compileExpr(stmt.Init), ptr)
			}
		}
	}
}

func (c *Compiler) compileFunc(def *FuncDefStmt) *ir.Func {
	Assert(def.Name != nil, "type checker has to handle this")
	Assert(def.Name.Name != "", "name can't be empty")
	name := def.Name.Name
	typ := c.resolveType(def.Type)

	params := make([]*ir.Param, 0)

	for _, def := range def.Params {
		typ2 := c.resolveType(def.Type)
		param := ir.NewParam(def.Name.Name, typ2)
		params = append(params, param)

		// TODO merge with compileVarDef function
		c.variables[def.Name.Name] = param
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
		f.Linkage = enum.LinkageExternal
	}

	return f
}

func (c *Compiler) compileStructDef(def *StructDefStmt) {
	str := types.NewStruct()
	for _, field := range def.Fields {
		typ := c.resolveType(field.Type)
		str.Fields = append(str.Fields, typ)
	}
	td := c.module.NewTypeDef(def.Name.Name, str)
	c.types[def.Name.Name] = td
}

func (c *Compiler) compileStmt(stmt Stmt) {
	switch innerStmt := stmt.(type) {
	case *ReturnStmt:
		c.currentBlock.NewRet(c.compileExpr(innerStmt.Value))
	case *BreakStmt:
		c.currentBlock.NewBr(c.breakBlock)
	case *ContinueStmt:
		c.currentBlock.NewBr(c.continueBlock)
	case *ExprStmt:
		c.compileExpr(innerStmt.Expr)
	case *IfStmt:
		c.compileIfStmt(innerStmt)
	case *FuncDefStmt:
		c.compileFunc(innerStmt)
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
	if v, ok := c.variables[name]; ok {
		return v
	}
	panic("no such variable: " + name)
}

func (c *Compiler) call(expr *CallExpr) value.Value {
	b := c.currentBlock

	fun := c.getFunc(expr.Name.Name)

	args := make([]value.Value, 0)
	for _, e := range expr.Args {
		arg := c.compileExpr(e)
		args = append(args, arg)
	}

	return b.NewCall(fun, args...)
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

func (c *Compiler) compileExpr(expr Expr) value.Value {
	switch expr := expr.(type) {
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

	case *CallExpr:
		return c.call(expr)
	case *IdentExpr:
		v := c.getAlloca(expr.Name)
		return c.currentBlock.NewLoad(v.(*ir.InstAlloca).ElemType, v)
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
	c.variables = make(map[string]value.Value)
	c.initBuiltinTypes()
	c.initBuiltinFuncs()
}

func (c *Compiler) getFunc(name string) *ir.Func {
	if f, ok := c.funcs[name]; ok {
		return f
	}
	return nil
}