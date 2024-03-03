package main

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"os"
	"strconv"
	"strings"
)

type CompileScope struct {
	variables map[string]value.Value
	Parent    *CompileScope
	Children  []*CompileScope
	funcs     map[string]*ir.Func
	// TODO Shorthands map[string]value.Value // Like this.a -> just a if it is not defined in the scope
}

func (s *CompileScope) Lookup(name string) value.Value {
	if s.variables[name] != nil {
		return s.variables[name]
	}

	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}

	panic("No variable found: " + name)
}

func (s *CompileScope) Define(name string, v value.Value) {
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
	currentScope  *CompileScope
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
		switch stmt := stmt.(type) {
		case *StructDefStmt:
			c.compileStructDef(stmt)
		}
	}

	for _, stmt := range c.pkg.Stmts {
		switch stmt := stmt.(type) {
		case *FuncDefStmt:
			c.compileFunc(stmt, false, nil, true)
		case *ImplStmt:
			thisType := c.resolveType(stmt.Type)
			for _, s := range stmt.Stmts {
				switch implStmt := s.(type) {
				case *FuncDefStmt:
					implStmt.ImplOf = stmt
					c.compileFunc(implStmt, true, thisType, true)
				}
			}

		}
	}

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
	c.types["i8"] = types.I8
	c.types["i16"] = types.I16
	c.types["i32"] = types.I32
	c.types["i64"] = types.I64
	c.types["bool"] = types.I1
	c.types["char"] = types.I8
	c.types["f16"] = types.Half
	c.types["f32"] = types.Float
	c.types["f64"] = types.Double

	stringStruct := types.NewStruct(types.NewPointer(types.I8), types.I64, types.I64)
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

	printfInternal := module.NewFunc("printf_internal", types.I32, ir.NewParam("", types.NewPointer(c.types["char"])))
	printfInternal.Sig.Variadic = true
	printfInternal.Linkage = enum.LinkageExternal
	c.funcs["printf_internal"] = printfInternal

	returnParam := ir.NewParam("", types.NewPointer(c.types["string"]))
	returnParam.Attrs = make([]ir.ParamAttribute, 0)
	returnParam.Attrs = append(returnParam.Attrs, ir.SRet{Typ: c.types["string"]})
	makeString := module.NewFunc("make_string", c.types["void"], returnParam, ir.NewParam("buffer", types.NewPointer(types.I8)))
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

	memcpyInternal := module.NewFunc("memcpy_internal", types.NewPointer(types.I8),
		ir.NewParam("dest", types.NewPointer(types.I8)),
		ir.NewParam("src", types.NewPointer(types.I8)),
		ir.NewParam("sizeof_i64", types.I64),
	)
	memcpyInternal.Linkage = enum.LinkageExternal
	c.funcs["memcpy_internal"] = memcpyInternal

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

func (c *Compiler) isNull(r value.Value) bool {
	_, ok := r.(*constant.Null)
	return ok
}
func (c *Compiler) compileVarDef(stmt *VarDefStmt) {
	Assert(c.currentScope != nil, "un-initialized scope")
	Assert(c.currentBlock != nil, "currentBlock cannot be nil")
	typ := c.resolveType(stmt.Type)
	c.currentType = typ
	ptr := c.currentBlock.NewAlloca(typ)
	c.currentAlloc = ptr
	c.Define(stmt.Name.Name, ptr)

	if stmt.Init != nil {
		right := c.compileExpr(stmt.Init)

		if !c.isAssignable(stmt.Context()) {
			// do memcpy from returned alloca
			c.memcpyInternal(ptr, right)
			return
		}

		switch right.Type().(type) {
		case *types.ArrayType:
			c.currentBlock.NewStore(right, ptr)

		case *types.PointerType:
			if !typ.Equal(right.Type()) || c.isNull(right) {
				cast := c.currentBlock.NewBitCast(right, typ)
				c.currentBlock.NewStore(cast, ptr)
			} else {
				c.currentBlock.NewStore(right, ptr)
			}
		case *types.FloatType:
			c.currentBlock.NewStore(right, ptr)
		default:
			c.currentBlock.NewStore(right, ptr)
		}
	}
}

func (c *Compiler) compileFunc(def *FuncDefStmt, isMethod bool, thisType types.Type, onlyDeclare bool) *ir.Func {
	Assert(def.Name != nil, "type checker has to handle this")
	Assert(def.Name.Name != "", "name can't be empty")
	name := def.Name.Name

	if isMethod {
		name = fmt.Sprintf("%s.%s", def.ImplOf.Type.(*IdentExpr).Name, name)
	}

	c.currentScope = &CompileScope{
		variables: make(map[string]value.Value),
		Parent:    c.currentScope,
		Children:  make([]*CompileScope, 0),
	}
	defer func() { c.currentScope = c.currentScope.Parent }()
	typ := c.resolveType(def.Type)

	params := make([]*ir.Param, 0)

	if t, ok := typ.(*types.StructType); ok {
		// struct types should be passed as parameter by the llvm standards
		returnParam := ir.NewParam("returnVal", types.NewPointer(t))
		returnParam.Attrs = make([]ir.ParamAttribute, 0)
		returnParam.Attrs = append(returnParam.Attrs, ir.SRet{Typ: t})
		returnParam.Attrs = append(returnParam.Attrs, enum.ParamAttrNoAlias)
		params = append(params, returnParam)
		typ = types.Void
	}

	if isMethod {
		param := ir.NewParam("this", types.NewPointer(thisType))
		params = append(params, param)
		c.Define("this", param)
	}

	for _, def := range def.Params {
		Assert(def.Name.Name != "this", "param name cannot be this")
		typ2 := c.resolveType(def.Type)
		var param *ir.Param
		if t2, ok := typ2.(*types.StructType); ok {
			param = ir.NewParam(def.Name.Name, types.NewPointer(t2))
			param.Attrs = make([]ir.ParamAttribute, 0)
			param.Attrs = append(param.Attrs, ir.Byval{Typ: t2})
		} else {
			param = ir.NewParam(def.Name.Name, typ2)
		}
		params = append(params, param)

		// TODO merge with compileVarDef function
		c.Define(def.Name.Name, param)
	}
	var f *ir.Func
	if onlyDeclare {
		f = c.module.NewFunc(name, typ, params...)
		c.funcs[name] = f
	} else {
		f = c.funcs[name]
	}
	c.currentFunc = f
	if !def.IsExternal {
		if !onlyDeclare {
			block := f.NewBlock(entryBlock)
			c.currentBlock = block

			if def.Context().returnsStruct {
				def.Context().returnParam = f.Params[0]
			}

			for _, innerStmt := range def.Body.Stmts {
				c.compileStmt(innerStmt)
			}

			if _, ok := typ.(*types.VoidType); ok && block.Term == nil {
				block.NewRet(nil)
			}

		}
	} else {
		Assert(!isMethod, "Methods cannot be external")
		f.Linkage = enum.LinkageExternal
	}

	return f
}

func (c *Compiler) compileStructDef(def *StructDefStmt) {

	if c.types[def.Name.Name] != nil {
		return
	}

	str := types.NewStruct()
	td := c.module.NewTypeDef(def.Name.Name, str)
	c.types[def.Name.Name] = td
	for i, field := range def.Fields {
		typ := c.resolveType(field.Type)
		str.Fields = append(str.Fields, typ)
		c.typesIndexMap[fieldIndexKey{field.Name.Name, def.Name.Name}] = i
	}
}

func (c *Compiler) compileStmt(stmt Stmt) {
	switch innerStmt := stmt.(type) {
	case *ReturnStmt:
		if innerStmt.Value == nil {
			c.currentBlock.NewRet(nil)
		} else {
			if innerStmt.Context().returnsStruct {
				param := innerStmt.Context().returnParam
				val := c.compileExpr(innerStmt.Value)
				c.currentBlock.NewStore(val, param)
				c.currentBlock.NewRet(nil)
			} else {
				c.currentBlock.NewRet(c.compileExpr(innerStmt.Value))
			}
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
		c.compileFunc(innerStmt, false, nil, false)
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
				c.compileFunc(implStmt, true, thisType, false)
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

func getFloatBit(t *types.FloatType) int64 {
	switch t.Kind {
	case types.FloatKindHalf:
		return 16
	case types.FloatKindFloat:
		return 32
	case types.FloatKindDouble:
		return 64
	default:
		panic("unhandled")
	}
}

func (c *Compiler) primitiveCasting(funcName string, expr *CallExpr) value.Value {
	toTyp := strings.Replace(funcName, "cast_", "", 1)
	arg := c.compileExpr(expr.Args[0])
	typ := c.resolveType(&IdentExpr{Name: toTyp})

	toBit := strings.Replace(strings.Replace(toTyp, "i", "", 1), "f", "", 1)
	i, err := strconv.ParseInt(toBit, 10, 64)
	AssertErr(err)

	if strings.HasPrefix(toTyp, "i") {
		switch t := arg.Type().(type) {
		case *types.IntType:
			if uint64(i) > t.BitSize {
				return c.currentBlock.NewSExt(arg, typ)
			} else if uint64(i) == t.BitSize {
				return arg
			} else {
				return c.currentBlock.NewTrunc(arg, typ)
			}
		case *types.FloatType:
			return c.currentBlock.NewFPToSI(arg, typ)

		default:
			panic("TODO")
		}
	} else if strings.HasPrefix(toTyp, "f") {
		switch t := arg.Type().(type) {
		case *types.IntType:
			return c.currentBlock.NewSIToFP(arg, typ)
		case *types.FloatType:
			floatBit := getFloatBit(t)
			if i > floatBit {
				return c.currentBlock.NewFPExt(arg, typ)
			} else if i == floatBit {
				return arg
			} else {
				return c.currentBlock.NewFPTrunc(arg, typ)
			}
		default:
			panic("TODO")
		}
	}

	panic("not implemented yet")

}

func (c *Compiler) call(expr *CallExpr) value.Value {
	b := c.currentBlock

	var isMethodCall bool
	var funcName string
	switch e := expr.Left.(type) {
	case *IdentExpr:
		funcName = e.Name
	case *SelectorExpr:
		isMethodCall = true
		funcName = e.Ident.(*IdentExpr).Name
	default:
		panic("TODO unreachable call")
	}
	args := make([]value.Value, 0)

	if expr.MethodOf != "" && isMethodCall {
		funcName = expr.MethodOf + "." + funcName
		args = append(args, c.getThisArg(expr.Left))
	}

	var returnVal value.Value = nil
	if !expr.TypeCast {
		fun := c.getFunc(funcName)

		if len(fun.Params) > 0 {
			param := fun.Params[0]
			if len(param.Attrs) > 0 {
				if _, ok := param.Attrs[0].(ir.SRet); ok {
					alloc := c.currentBlock.NewAlloca(param.Type().(*types.PointerType).ElemType)
					args = append(args, alloc)
					returnVal = alloc
				}
			}
		}

		for i, e := range expr.Args {
			arg := c.compileExpr(e)
			if funcName == "printf_internal" {
				switch t := arg.Type().(type) {
				case *types.FloatType:
					if t.Kind != types.FloatKindDouble {
						// think a way better
						arg = b.NewFPExt(arg, types.Double)
					}
				}

			}

			if i == 0 {
				idx := 0
				if returnVal != nil {
					idx += 1
				}
				param := fun.Params[idx]
				if len(param.Attrs) > 0 {

					if _, ok := param.Attrs[0].(ir.Byval); ok {
						if _, ok := arg.Type().(*types.PointerType); !ok {
							alloca := c.currentBlock.NewAlloca(arg.Type())
							c.memcpyInternal(alloca, arg.(*ir.InstLoad).Src)
							arg = alloca
						}
					}
				}
			}

			args = append(args, arg)
		}

		callRes := b.NewCall(fun, args...)
		if returnVal != nil {
			return returnVal
		}
		return callRes
	} else {
		return c.primitiveCasting(funcName, expr)
	}
}

func (c *Compiler) resolveFieldIndex(name, typ string) (int, error) {
	if index, ok := c.typesIndexMap[fieldIndexKey{name, typ}]; ok {
		return index, nil
	}

	return -1, fmt.Errorf("unknown field index: %v for %s", name, typ)
}

func (c *Compiler) callSizeOf(t types.Type) value.Value {
	switch t := t.(type) {
	case *types.StructType, *types.ArrayType, *types.IntType, *types.FloatType:
		alloca := c.currentBlock.NewAlloca(t)
		endPtr := c.currentBlock.NewGetElementPtr(t, alloca, constant.NewInt(types.I32, 1))
		ptrToInt := c.currentBlock.NewPtrToInt(alloca, types.I64)
		endPtrInt := c.currentBlock.NewPtrToInt(endPtr, types.I64)
		sizeInBytes := generateOperation(c.currentBlock, endPtrInt, ptrToInt, Sub)
		return sizeInBytes
	default:
		panic("implement sizeof case")
	}
}

// sizeOfCompilationTime returns the size of
func sizeof(p types.Type) (int64 uint64) {
	switch p := p.(type) {
	case *types.IntType:
		return p.BitSize / 8
	case *types.FloatType:
		return 8
	case *types.PointerType:
		return 8
	case *types.StructType:
		var s uint64 = 0
		for _, field := range p.Fields {
			s += sizeof(field)
		}
		return s
	case *types.ArrayType:
		return sizeof(p.ElemType) * p.Len
	default:
		panic("TODO unreachable sizeof")
	}
}

func (c *Compiler) newArray(expr *ArrayLitExpr, typ types.Type) value.Value {

	var alloc = c.currentBlock.NewAlloca(typ)

	elems := expr.Elems
	for i, elem := range elems {
		e := c.compileExpr(elem)
		zero := constant.NewInt(types.I64, 0)
		index := constant.NewInt(types.I64, int64(i))
		gep := c.currentBlock.NewGetElementPtr(typ, alloc, zero, index)
		gep.InBounds = true
		if ee, ok := e.(*ir.InstAlloca); ok {
			c.memcpyInternal(gep, ee)
		} else {
			c.currentBlock.NewStore(e, gep)
		}
	}

	return alloc
}

func (c *Compiler) getDefaultValue(p types.Type) constant.Constant {

	switch t := p.(type) {
	case *types.IntType:
		return constant.NewInt(t, 0)
	case *types.FloatType:
		return constant.NewFloat(t, 0)
	case *types.PointerType:
		return constant.NewNull(t)
	case *types.StructType:
		obj := constant.NewStruct(t)
		var fields []constant.Constant
		for _, f := range t.Fields {
			fields = append(fields, c.getDefaultValue(f))
		}
		obj.Fields = fields
		return obj
	default:
		panic("not implemented default value case!")
	}

	return nil
}

func (c *Compiler) resolveFieldType(typ *types.StructType, index int64) types.Type {
	return typ.Fields[index]
}

func (c *Compiler) newObject(expr *ObjectLitExpr, alloc value.Value) value.Value {
	typ := c.resolveType(expr.Type)
	t, ok := typ.(*types.StructType)
	Assert(ok, "Object without struct type!")
	for _, kv := range expr.KeyValue {
		key, ok := kv.Key.(*IdentExpr)
		Assert(ok, "Key must be a identifier")
		fieldIndex, err := c.resolveFieldIndex(key.Name, typ.Name())
		AssertErr(err)
		v := c.compileExpr(kv.Value)
		if c.isNull(v) {
			gep := c.currentBlock.NewGetElementPtr(typ, alloc, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, int64(fieldIndex)))
			v2 := c.currentBlock.NewBitCast(v, c.resolveFieldType(t, int64(fieldIndex)))
			c.currentBlock.NewStore(v2, gep)
		} else {
			gep := c.currentBlock.NewGetElementPtr(typ, alloc, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, int64(fieldIndex)))
			c.currentBlock.NewStore(v, gep)
		}
	}

	return alloc
}

func (c *Compiler) getIndexedType(load *ir.InstLoad, idx int) types.Type {
	switch l := load.ElemType.(type) {
	case *types.StructType:
		return l.Fields[idx]
	default:
		panic("Unreachable indexed type")
	}
}

func (c *Compiler) getThisArg(expr Expr) value.Value {
	switch expr := expr.(type) {
	case *SelectorExpr:
		sel := c.compileExpr(expr.Selector)
		var elemType types.Type
		var returnVal value.Value
		switch s := sel.(type) {
		case *ir.InstLoad:
			elemType = s.Src.(*ir.InstAlloca).ElemType
			returnVal = s
		case *ir.InstCall:
			elemType = s.Typ.(*types.PointerType).ElemType
			returnVal = s
		default:
			panic("implement me getThisArg()")
		}
		_, ok2 := elemType.(*types.PointerType)
		if ok2 {
			return sel
		}
		return returnVal
	case *IdentExpr:
		return c.compileExpr(expr)
	default:
		panic("Unreachable")
	}
}

func (c *Compiler) memcpyInternal(dest, src value.Value) value.Value {
	f := c.getFunc("memcpy_internal")
	sizeInBytes := c.callSizeOf(src.Type().(*types.PointerType).ElemType)
	a1 := c.currentBlock.NewBitCast(dest, types.NewPointer(types.I8))
	a2 := c.currentBlock.NewBitCast(src, types.NewPointer(types.I8))
	return c.currentBlock.NewCall(f, a1, a2, sizeInBytes)
}

func (c *Compiler) mallocInternal(size uint64) value.Value {
	callExpr := &CallExpr{
		Left: &IdentExpr{Name: "malloc_internal"},
		Args: []Expr{
			&NumberExpr{
				Value: int64(size),
				ctx: &VarAssignCtx{
					parent:       nil,
					expectedType: "i64",
				},
			},
		},
		end:      Pos{},
		MethodOf: "",
		TypeCast: false,
	}

	return c.call(callExpr)
}

func (c *Compiler) getSizeOf(expr Expr) uint64 {
	switch expr := expr.(type) {
	case *ObjectLitExpr:
		return sizeof(c.resolveType(expr.Type))
	case *IdentExpr:
		val := c.Lookup(expr.Name)
		return sizeof(val.Type())
	default:
		panic("TODO")
	}
}

func (c *Compiler) isAssignable(ctx *VarAssignCtx) bool {

	if ctx.ExpectedType() == "string" {
		return false
	}

	if ctx.isArray {
		return false
	}

	if ctx.isStruct {
		return false
	}

	return true

}

func (c *Compiler) compileExpr(expr Expr) value.Value {
	switch expr := expr.(type) {
	case *NilExpr:
		null := constant.NewNull(types.NewPointer(types.I8))
		return null
	case *CharExpr:
		return constant.NewInt(types.I8, int64(expr.Unquoted))
	case *ArrayLitExpr:
		var arrayType *ArrayTypeExpr
		switch ctx := expr.GetCtx().(type) {
		case *VarAssignCtx:
			arrayType = &ArrayTypeExpr{
				Type: &IdentExpr{Name: ctx.extractedType},
				Size: &NumberExpr{Value: int64(ctx.arraySize)},
				end:  Pos{},
			}
		default:
			panic("not implemented ctx for array lit")
		}
		typ := c.resolveType(arrayType)
		return c.newArray(expr, typ)
	case *BinaryExpr:
		left := c.compileExpr(expr.Left)
		right := c.compileExpr(expr.Right)
		var v = left
		switch e := expr.Left.(type) {
		case *UnaryExpr:
			if e.Op != Sizeof {
				v = c.currentBlock.NewLoad(types.I32, v)
			}
		}

		return generateOperation(c.currentBlock, v, right, expr.Op)
	case *NumberExpr:
		ct := expr.GetCtx()
		ctx, ok := ct.(ExpectedTypeCtx)
		Assert(ok, "Unimplemented number type")
		size := getBitSize(ctx.ExpectedType())
		if size == -1 {
			size = 32 // default
		}
		return constant.NewInt(types.NewInt(uint64(size)), expr.Value)
	case *FloatExpr:
		ct := expr.GetCtx()
		ctx, ok := ct.(ExpectedTypeCtx)
		Assert(ok, "Unimplemented ctx type float")
		size := getBitSize(ctx.ExpectedType())
		if size == -1 {
			size = 64 // default
		}
		switch size {
		case 16:
			return constant.NewFloat(types.Half, expr.Value)
		case 32:
			return constant.NewFloat(types.Float, expr.Value)
		case 64:
			return constant.NewFloat(types.Double, expr.Value)
		default:
			panic("unhandled")

		}
	case *StringExpr:
		v := expr.Unquoted
		v2 := constant.NewCharArrayFromString(v)
		strPtr := c.module.NewGlobalDef("", v2)
		strPtr.Immutable = true
		strPtr.UnnamedAddr = enum.UnnamedAddrUnnamedAddr
		var expectedType = "string"
		ctx, ok := expr.GetCtx().(ExpectedTypeCtx)
		Assert(ok, "Unimplemented String Ctx")
		expectedType = ctx.ExpectedType()
		var zero = constant.NewInt(types.I64, 0)
		gep := constant.NewGetElementPtr(v2.Typ, strPtr, zero, zero)
		gep.InBounds = true
		if expectedType == "pointer<char>" {
			return gep
		}
		alloc := c.currentBlock.NewAlloca(c.types["string"])
		alloc.Align = 8
		c.currentBlock.NewCall(c.getFunc("make_string"), alloc, gep)
		return c.currentBlock.NewLoad(c.types["string"], alloc)
	case *ObjectLitExpr:
		typ := c.resolveType(expr.Type)
		t, ok := typ.(*types.StructType)
		Assert(ok, "Object without struct type!")
		alloc := c.currentBlock.NewAlloca(t)
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
			switch e := expr.Right.(type) {
			case *ObjectLitExpr:
				size := c.getSizeOf(expr.Right)
				allocation := c.mallocInternal(size)
				x := c.currentBlock.NewBitCast(allocation, types.NewPointer(c.resolveType(e.Type)))
				val := c.newObject(e, x)
				return val
			case *IdentExpr:
				return c.getAlloca(e.Name)
			default:
				panic("Unhandled assignment case")
			}
		case Sizeof:
			// TODO constant expressions must, we handle as statically to save day for now
			resolvedType := c.resolveType(expr.Right)
			return c.callSizeOf(resolvedType)
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
		if c.isAssignable(expr.Context()) {
			c.currentBlock.NewStore(right, load.Src)
		} else {
			c.memcpyInternal(load.Src, right)
		}
		return left
	case *IndexExpr:
		left := c.compileExpr(expr.Left)
		index := c.compileExpr(expr.Index)
		zero := constant.NewInt(types.I64, 0)
		leftInstLoad, ok := left.(*ir.InstLoad)
		Assert(ok, "IndexExpr left operand is invalid.")

		switch t := leftInstLoad.ElemType.(type) {
		case *types.ArrayType:
			gep := c.currentBlock.NewGetElementPtr(leftInstLoad.ElemType, leftInstLoad.Src, zero, index)
			return c.currentBlock.NewLoad(t.ElemType, gep)
		case *types.StructType:
			if t.Name() == "string" {
				bufferGep := c.currentBlock.NewGetElementPtr(leftInstLoad.ElemType, leftInstLoad.Src, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, 0))
				loadCharPointer := c.currentBlock.NewLoad(types.NewPointer(types.I8), bufferGep)
				gep := c.currentBlock.NewGetElementPtr(types.I8, loadCharPointer, index)
				return c.currentBlock.NewLoad(types.I8, gep)
			} else {
				panic("Index operation cannot be used with non-array type")
			}
		case *types.PointerType:
			gep := c.currentBlock.NewGetElementPtr(t.ElemType, leftInstLoad, index)
			return c.currentBlock.NewLoad(t.ElemType, gep)
		default:
			panic("Index operation cannot be used with non-array type")
		}

	case *SelectorExpr:
		sel := c.compileExpr(expr.Selector)
		selLoad, ok := sel.(*ir.InstLoad)
		Assert(ok, "Selector invalid")
		if s, ok := selLoad.ElemType.(*types.PointerType); ok {
			selLoad = c.currentBlock.NewLoad(s.ElemType, selLoad)
		}
		zero := constant.NewInt(types.I32, 0)
		typName := sel.Type().Name()
		if t, ok := sel.Type().(*types.PointerType); ok {
			typName = t.ElemType.Name()
		}
		fieldIndex, err := c.resolveFieldIndex(expr.Ident.(*IdentExpr).Name, typName)
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
	_, ok1 := value1.Type().(*types.FloatType)
	_, ok2 := value2.Type().(*types.FloatType)
	switch op {
	case Add:
		if ok1 || ok2 {
			return block.NewFAdd(value1, value2)
		}
		return block.NewAdd(value1, value2)
	case Sub:
		if ok1 || ok2 {
			return block.NewFSub(value1, value2)
		}
		return block.NewSub(value1, value2)
	case Mul:
		if ok1 || ok2 {
			return block.NewFMul(value1, value2)
		}
		return block.NewMul(value1, value2)
	case Div:
		if ok1 || ok2 {
			return block.NewFDiv(value1, value2)
		}
		return block.NewUDiv(value1, value2)
	case Mod:
		if ok1 || ok2 {
			return block.NewFRem(value1, value2)
		}
		return block.NewURem(value1, value2)
	case Eq:
		if ok1 || ok2 {
			return block.NewFCmp(enum.FPredOEQ, value1, value2)
		}
		return block.NewICmp(enum.IPredEQ, value1, value2)
	case Neq:
		if ok1 || ok2 {
			return block.NewFCmp(enum.FPredONE, value1, value2)
		}
		return block.NewICmp(enum.IPredNE, value1, value2)
	case Gt:
		if ok1 || ok2 {
			return block.NewFCmp(enum.FPredOGT, value1, value2)
		}
		return block.NewICmp(enum.IPredSGT, value1, value2)
	case Gte:
		if ok1 || ok2 {
			return block.NewFCmp(enum.FPredOGE, value1, value2)
		}
		return block.NewICmp(enum.IPredSGE, value1, value2)
	case Lt:
		if ok1 || ok2 {
			return block.NewFCmp(enum.FPredOLT, value1, value2)
		}
		return block.NewICmp(enum.IPredSLT, value1, value2)
	case Lte:
		if ok1 || ok2 {
			return block.NewFCmp(enum.FPredOLE, value1, value2)
		}
		return block.NewICmp(enum.IPredSLE, value1, value2)
	case And:
		return block.NewAnd(value1, value2)
	case Or:
		return block.NewOr(value1, value2)
	default:
		panic("unhandled default case")
	}

}

func (c *Compiler) init() {
	Assert(c.module != nil, "module not initialized")

	c.currentScope = &CompileScope{
		variables: make(map[string]value.Value),
		Parent:    nil,
		Children:  make([]*CompileScope, 0),
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
