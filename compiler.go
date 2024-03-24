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
	module       *ir.Module
	currentScope *CompileScope
	currentBlock *ir.Block
	currentFunc  *ir.Func
	currentType  types.Type
	currentAlloc *ir.InstAlloca

	funcs         map[string]*ir.Func
	types         map[string]types.Type
	typesIndexMap map[fieldIndexKey]int
	globals       map[string]value.Value
	pkg           *Package
	sequence      int
}

func (c *Compiler) GetSequence() string {
	c.sequence += 1
	return fmt.Sprintf("%d", c.sequence)
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

	sliceX := types.NewStruct(
		types.NewPointer(types.NewPointer(types.I8)),
		types.NewInt(64),
		types.NewInt(64),
	)
	sliceDef := c.module.NewTypeDef("slice", sliceX)
	c.types["slice"] = sliceDef

	stringStruct := types.NewStruct(types.NewPointer(types.I8), types.I64, types.I64)
	def := c.module.NewTypeDef("string", stringStruct)
	c.types["string"] = def
}

func (c *Compiler) NewPassByValueParameter(name string, typ types.Type) *ir.Param {
	param := ir.NewParam(name, types.NewPointer(typ))
	param.Attrs = make([]ir.ParamAttribute, 0)
	param.Attrs = append(param.Attrs, PassByValue{Typ: typ})
	return param
}

func (c *Compiler) NewReturnParameter(name string, typ types.Type) *ir.Param {
	param := ir.NewParam(name, types.NewPointer(typ))
	param.Attrs = make([]ir.ParamAttribute, 0)
	param.Attrs = append(param.Attrs, ir.SRet{Typ: typ})
	param.Attrs = append(param.Attrs, enum.ParamAttrNoAlias)
	return param
}
func (c *Compiler) initBuiltinFuncs() {
	module := c.module

	if c.funcs == nil {
		c.funcs = make(map[string]*ir.Func)
	}

	printfInternal := module.NewFunc("printf_internal", types.I32, ir.NewParam("", types.NewPointer(c.types["char"])))
	printfInternal.Sig.Variadic = true
	printfInternal.Linkage = enum.LinkageExternal
	c.funcs["printf_internal"] = printfInternal

	makeString := module.NewFunc("make_string", c.types["void"], c.NewReturnParameter("", c.types["string"]), ir.NewParam("buffer", types.NewPointer(types.I8)))
	makeString.Linkage = enum.LinkageExternal
	c.funcs["make_string"] = makeString

	concatStrings := module.NewFunc("concat_strings", c.types["void"], c.NewReturnParameter("", c.types["string"]),
		c.NewPassByValueParameter("strA", c.types["string"]),
		c.NewPassByValueParameter("strB", c.types["string"]))
	concatStrings.Linkage = enum.LinkageExternal
	c.funcs["concat_strings"] = concatStrings

	concatCharAndString := module.NewFunc("concat_char_and_string", c.types["void"],
		c.NewReturnParameter("", c.types["string"]), ir.NewParam("", c.types["char"]),
		c.NewPassByValueParameter("", c.types["string"]))
	concatCharAndString.Linkage = enum.LinkageExternal
	c.funcs["concat_char_and_string"] = concatCharAndString

	concatStringAndChar := module.NewFunc("concat_string_and_char", c.types["void"],
		c.NewReturnParameter("", c.types["string"]), c.NewPassByValueParameter("", c.types["string"]),
		ir.NewParam("", c.types["char"]))
	concatStringAndChar.Linkage = enum.LinkageExternal
	c.funcs["concat_string_and_char"] = concatStringAndChar

	concatCharAndChar := module.NewFunc("concat_char_and_char", c.types["void"],
		c.NewReturnParameter("", c.types["string"]),
		ir.NewParam("", c.types["char"]),
		ir.NewParam("", c.types["char"]))
	concatCharAndChar.Linkage = enum.LinkageExternal
	c.funcs["concat_char_and_char"] = concatCharAndChar

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

	compareString := module.NewFunc("compare_string", types.I1,
		c.NewPassByValueParameter("", c.types["string"]),
		c.NewPassByValueParameter("", c.types["string"]))
	c.funcs["compare_string"] = compareString

	makeSlice := module.NewFunc("make_slice", types.Void,
		c.NewReturnParameter("returnVal", c.types["slice"]))
	makeSlice.Linkage = enum.LinkageExternal
	c.funcs["make_slice"] = makeSlice

	accessSliceIndex := module.NewFunc("access_slice_data",
		types.I64, c.NewPassByValueParameter("slice", c.types["slice"]), ir.NewParam("", types.I64))
	accessSliceIndex.Linkage = enum.LinkageExternal
	c.funcs["access_slice_data"] = accessSliceIndex

	accessSliceIndexP := module.NewFunc("access_slice_datap",
		types.I64, c.NewPassByValueParameter("slice", c.types["slice"]), ir.NewParam("", types.I64))
	accessSliceIndex.Linkage = enum.LinkageExternal
	c.funcs["access_slice_datap"] = accessSliceIndexP

	appendSliceData := module.NewFunc("append_slice_data",
		types.Void, ir.NewParam("", types.NewPointer(c.types["slice"])), ir.NewParam("", types.NewPointer(types.I8)))
	appendSliceData.Linkage = enum.LinkageExternal
	c.funcs["append"] = appendSliceData

	appendSliceDatap := module.NewFunc("append_slice_datap",
		types.Void, ir.NewParam("", types.NewPointer(c.types["slice"])), ir.NewParam("", types.NewPointer(types.I8)))
	appendSliceData.Linkage = enum.LinkageExternal
	c.funcs["appendp"] = appendSliceDatap

	lenSlice := module.NewFunc("len_slice", types.I64, c.NewPassByValueParameter("", c.types["slice"]))
	lenSlice.Linkage = enum.LinkageExternal
	c.funcs["len"] = lenSlice

	printStrFn := module.NewFunc("__print_str__", types.Void, c.NewPassByValueParameter("str", c.types["string"]))
	printStrFn.Linkage = enum.LinkageExternal
	c.funcs["__print_str__"] = printStrFn

	printI8Fn := module.NewFunc("__print_i8__", types.Void, ir.NewParam("", c.types["i8"]))
	printI8Fn.Linkage = enum.LinkageExternal
	c.funcs["__print_i8__"] = printI8Fn

	printI16Fn := module.NewFunc("__print_i16__", types.Void, ir.NewParam("", c.types["i16"]))
	printI16Fn.Linkage = enum.LinkageExternal
	c.funcs["__print_i16__"] = printI16Fn

	printI32Fn := module.NewFunc("__print_i32__", types.Void, ir.NewParam("", c.types["i32"]))
	printI32Fn.Linkage = enum.LinkageExternal
	c.funcs["__print_i32__"] = printI32Fn

	printI64Fn := module.NewFunc("__print_i64__", types.Void, ir.NewParam("", c.types["i64"]))
	printI64Fn.Linkage = enum.LinkageExternal
	c.funcs["__print_i64__"] = printI64Fn

	printF16Fn := module.NewFunc("__print_f16__", types.Void, ir.NewParam("", c.types["f16"]))
	printF16Fn.Linkage = enum.LinkageExternal
	c.funcs["__print_f16__"] = printF16Fn

	printF32Fn := module.NewFunc("__print_f32__", types.Void, ir.NewParam("", c.types["f32"]))
	printF32Fn.Linkage = enum.LinkageExternal
	c.funcs["__print_f32__"] = printF32Fn

	printF64Fn := module.NewFunc("__print_f64__", types.Void, ir.NewParam("", c.types["f64"]))
	printF64Fn.Linkage = enum.LinkageExternal
	c.funcs["__print_f64__"] = printF64Fn

	printCharFn := module.NewFunc("__print_char__", types.Void, ir.NewParam("_", c.types["char"]))
	printCharFn.Linkage = enum.LinkageExternal
	c.funcs["__print_char__"] = printCharFn

	printCharPFn := module.NewFunc("__print_charp__", types.Void, ir.NewParam("_", types.NewPointer(c.types["char"])))
	printCharPFn.Linkage = enum.LinkageExternal
	c.funcs["__print_charp__"] = printCharPFn

	printLnFn := module.NewFunc("__print_ln__", types.Void)
	printLnFn.Linkage = enum.LinkageExternal
	c.funcs["__print_ln__"] = printLnFn

	printBoolFn := module.NewFunc("__print__bool__", types.Void, ir.NewParam("_", c.types["bool"]))
	printBoolFn.Linkage = enum.LinkageExternal
	c.funcs["__print_bool__"] = printBoolFn
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
		if expr.Size.(*NumberExpr).Value == -1 {
			return c.types["slice"]
		}
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

		if !c.isStorable(stmt.Context()) {
			// do memcpy from returned alloca
			switch r := right.(type) {
			case *ir.InstLoad:
				c.memcpyInternal(ptr, r.Src)
			default:
				c.memcpyInternal(ptr, right)
			}
			return
		}

		switch right.Type().(type) {
		case *types.ArrayType:
			c.currentBlock.NewStore(right, ptr)

		case *types.PointerType:
			if c.isNull(right) {
				cast := c.currentBlock.NewBitCast(right, typ)
				c.currentBlock.NewStore(cast, ptr)
			} else if !typ.Equal(right.Type()) {
				load := c.currentBlock.NewLoad(types.NewPointer(typ.(*types.PointerType).ElemType), right)
				cast := c.currentBlock.NewBitCast(load, typ)
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

	if isMethod {
		param := ir.NewParam("this", types.NewPointer(thisType))
		params = append(params, param)
		c.Define("this", param)
	}

	if t, ok := typ.(*types.StructType); ok {
		// struct types should be passed as parameter by the llvm standards
		returnParam := ir.NewParam("returnVal", types.NewPointer(t))
		returnParam.Attrs = make([]ir.ParamAttribute, 0)
		returnParam.Attrs = append(returnParam.Attrs, ir.SRet{Typ: t})
		returnParam.Attrs = append(returnParam.Attrs, enum.ParamAttrNoAlias)
		params = append(params, returnParam)
		typ = types.Void
	}

	for _, def := range def.Params {
		Assert(def.Name.Name != "this", "param name cannot be this")
		typ2 := c.resolveType(def.Type)
		var param *ir.Param
		if t2, ok := typ2.(*types.StructType); ok {
			param = ir.NewParam(def.Name.Name, types.NewPointer(t2))
			param.Attrs = make([]ir.ParamAttribute, 0)
			param.Attrs = append(param.Attrs, PassByValue{Typ: t2})
		} else {
			param = ir.NewParam(def.Name.Name, typ2)
		}
		params = append(params, param)

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

			if def.Context().returnsStruct || def.Context().ExpectedType() == "string" {
				var returnParam *ir.Param
				for _, p := range f.Params {
					if p.Name() == "returnVal" {
						returnParam = p
					}
				}
				def.Context().returnParam = returnParam
			}

			for _, innerStmt := range def.Body.Stmts {
				c.compileStmt(innerStmt)
			}

			if _, ok := typ.(*types.VoidType); ok && c.currentBlock.Term == nil {
				c.currentBlock.NewRet(nil)
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
			if innerStmt.Context().returnsStruct || innerStmt.Context().ExpectedType() == "string" {
				param := innerStmt.Context().returnParam
				val := c.compileExpr(innerStmt.Value)
				if _, ok := val.(*ir.InstAlloca); ok {
					c.memcpyInternal(param, val)
				} else {
					c.currentBlock.NewStore(val, param)
				}
				c.currentBlock.NewRet(nil)
			} else {
				val := c.compileExpr(innerStmt.Value)
				// TODO pass type of pointer and return directly instead bit cast
				if c.isNull(val) {
					name := innerStmt.Context().returnTypeSym.TypeName()
					typ := c.resolveType(&RefTypeExpr{Expr: &IdentExpr{Name: name}})
					cast := c.currentBlock.NewBitCast(val, typ)
					c.currentBlock.NewRet(cast)
					return
				}
				c.currentBlock.NewRet(val)
			}
		}
	case *BreakStmt:
		c.currentBlock.NewBr(innerStmt.ctx.breakBlock)
	case *ContinueStmt:
		c.currentBlock.NewBr(innerStmt.ctx.continueBlock)
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
		seenBreakingStmt := false
		for _, innerStmt := range innerStmt.Stmts {
			if seenBreakingStmt {
				break
			}
			c.compileStmt(innerStmt)
			switch innerStmt.(type) {
			case *BreakStmt, *ContinueStmt, *ReturnStmt:
				seenBreakingStmt = true
			}
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
	seq := c.GetSequence()
	cond := c.compileExpr(stmt.Cond)
	thenBlock := c.currentFunc.NewBlock("if_then_" + seq)
	continueBlock := c.currentFunc.NewBlock("if_continue_" + seq)
	var elseBlock *ir.Block = continueBlock
	if stmt.Else != nil {
		elseBlock = c.currentFunc.NewBlock("else_" + seq)
	}

	c.currentBlock.NewCondBr(cond, thenBlock, elseBlock)
	c.currentBlock = thenBlock
	c.currentScope = &CompileScope{
		variables: make(map[string]value.Value),
		Parent:    c.currentScope,
		Children:  make([]*CompileScope, 0),
		funcs:     make(map[string]*ir.Func),
	}
	c.compileStmt(stmt.Then)
	c.currentScope = c.currentScope.Parent

	if thenBlock.Term == nil {
		thenBlock.NewBr(continueBlock)
	}

	if c.currentBlock.Term == nil {
		c.currentBlock.NewBr(continueBlock)
	}

	if stmt.Else != nil {
		c.currentBlock = elseBlock
		c.currentScope = &CompileScope{
			variables: make(map[string]value.Value),
			Parent:    c.currentScope,
			Children:  make([]*CompileScope, 0),
			funcs:     make(map[string]*ir.Func),
		}
		c.compileStmt(stmt.Else)
		c.currentScope = c.currentScope.Parent
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


	initBlock -> condBlock -> forBlock -> stepperBlock -> condBlock
							-> parentBlockContinue
*/

func (c *Compiler) compileForStmt(stmt *ForStmt) {

	c.currentScope = &CompileScope{
		variables: make(map[string]value.Value),
		Parent:    c.currentScope,
		Children:  make([]*CompileScope, 0),
		funcs:     make(map[string]*ir.Func),
	}

	defer func() { c.currentScope = c.currentScope.Parent }()
	seq := c.GetSequence()
	stmt.ctx.breakBlock = c.currentFunc.NewBlock("for_break_block_" + seq)
	initBlock := c.currentFunc.NewBlock("init_" + seq)
	forBlock := c.currentFunc.NewBlock("for_body_" + seq)
	condBlock := c.currentFunc.NewBlock("for_cond_" + seq)
	stepBlock := c.currentFunc.NewBlock("step_" + seq)

	stmt.ctx.forBlock = forBlock
	stmt.ctx.condBlock = condBlock
	stmt.ctx.stepBlock = stepBlock
	stmt.ctx.initBlock = initBlock
	stmt.ctx.continueBlock = condBlock

	c.currentBlock.NewBr(initBlock)
	c.currentBlock = initBlock
	if stmt.Init != nil {
		c.compileStmt(stmt.Init)
	}
	initBlock.NewBr(condBlock)

	c.currentBlock = condBlock
	condBlock.NewCondBr(c.compileExpr(stmt.Cond), forBlock, stmt.ctx.breakBlock)
	if stmt.Step != nil {
		c.currentBlock = stepBlock
		stmt.ctx.continueBlock = stepBlock
		c.compileExpr(stmt.Step)
		stepBlock.NewBr(condBlock)
	} else {
		stepBlock.NewBr(condBlock)
	}

	c.currentBlock = forBlock

	c.compileStmt(stmt.Body)
	if c.currentBlock != forBlock && c.currentBlock.Term == nil {
		c.currentBlock.NewBr(stmt.ctx.continueBlock)
	}
	if forBlock.Term == nil {
		c.currentBlock.NewBr(stmt.ctx.continueBlock)
	}

	c.currentBlock = stmt.ctx.breakBlock
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

func (c *Compiler) _callAppend(p0, p1 value.Value) value.Value {
	fun := c.funcs["append"]

	if _, ok := p1.(*ir.InstAlloca); ok {
		fun = c.funcs["append"]
	} else {
		if _, ok := p1.Type().(*types.PointerType); ok {
			fun = c.funcs["appendp"]
		}
	}

	a1 := c.currentBlock.NewAlloca(p1.Type())
	if p11, ok := p1.(*ir.InstAlloca); ok {
		c.memcpyInternal(a1, p11)
	} else {
		c.currentBlock.NewStore(p1, a1)
	}

	c1 := c.currentBlock.NewBitCast(a1, types.NewPointer(types.I8))
	return c.currentBlock.NewCall(fun, p0.(*ir.InstLoad).Src, c1)
}

func (c *Compiler) callAppend(expr *CallExpr) value.Value {
	p0 := c.compileExpr(expr.Args[0])
	p1 := c.compileExpr(expr.Args[1])
	return c._callAppend(p0, p1)
}

func (c *Compiler) callPrint(expr *CallExpr, newline bool) value.Value {
	var val value.Value
	if len(expr.Args) == 1 {
		arg := c.compileExpr(expr.Args[0])
		switch t := arg.Type().(type) {
		case *types.IntType:
			if t.BitSize == 1 {
				fun := c.funcs["__print_bool__"]
				val = c.currentBlock.NewCall(fun, arg)
			} else if t.BitSize == 8 {
				// TODO handle by context type (char or i8)
				fun := c.funcs["__print_char__"]
				val = c.currentBlock.NewCall(fun, arg)
			} else if t.BitSize == 16 {
				fun := c.funcs["__print_i16__"]
				val = c.currentBlock.NewCall(fun, arg)
			} else if t.BitSize == 32 {
				fun := c.funcs["__print_i32__"]
				val = c.currentBlock.NewCall(fun, arg)
			} else if t.BitSize == 64 {
				fun := c.funcs["__print_i64__"]
				val = c.currentBlock.NewCall(fun, arg)
			} else {
				panic("could not find func for i" + strconv.FormatUint(t.BitSize, 10))
			}
		case *types.FloatType:
			if t.Kind == types.FloatKindHalf {
				val = c.currentBlock.NewCall(c.funcs["__print_f16__"], arg)
			} else if t.Kind == types.FloatKindFloat {
				val = c.currentBlock.NewCall(c.funcs["__print_f32__"], arg)
			} else if t.Kind == types.FloatKindDouble {
				val = c.currentBlock.NewCall(c.funcs["__print_f64__"], arg)
			} else {
				panic("could not find func for float")
			}
		case *types.StructType:
			if t.Name() == "string" {
				alloca := c.currentBlock.NewAlloca(arg.Type())
				c.memcpyInternal(alloca, arg.(*ir.InstLoad).Src)
				arg = alloca
				val = c.currentBlock.NewCall(c.funcs["__print_str__"], arg)
			} else {
				panic("could not find func for unknown struct type")
			}
		case *types.PointerType:
			if t.ElemType.Equal(c.types["char"]) {
				val = c.currentBlock.NewCall(c.funcs["__print_charp__"], arg)
			} else if t.ElemType.Name() == "string" {
				val = c.currentBlock.NewCall(c.funcs["__print_str__"], arg)
			} else {
				panic("unknown case")
			}
		default:
			panic("unknown print value")
		}
	}

	if newline {
		return c.currentBlock.NewCall(c.funcs["__print_ln__"])
	}

	return val
}

func (c *Compiler) customCall(expr *CallExpr) value.Value {
	var funcName string
	switch e := expr.Left.(type) {
	case *IdentExpr:
		funcName = e.Name
	default:
		panic("TODO handle custom call")
	}

	if funcName == "append" {
		return c.callAppend(expr)
	} else if funcName == "print" {
		return c.callPrint(expr, false)
	} else if funcName == "println" {
		return c.callPrint(expr, true)
	} else {
		panic("Unimplemented custom call")
	}
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
			id := 0
			if isMethodCall {
				id = 1
			}

			if id < len(fun.Params) {
				param := fun.Params[id]
				if len(param.Attrs) > 0 {
					if _, ok := param.Attrs[0].(ir.SRet); ok {
						alloc := c.currentBlock.NewAlloca(param.Type().(*types.PointerType).ElemType)
						args = append(args, alloc)
						returnVal = alloc
					}
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

			idx := i
			if returnVal != nil {
				idx += 1
			}
			if isMethodCall {
				idx += 1
			}

			if idx < len(fun.Params) {
				param := fun.Params[idx]
				if len(param.Attrs) > 0 {

					if _, ok := param.Attrs[0].(PassByValue); ok {
						if _, ok := arg.Type().(*types.PointerType); !ok {
							alloca := c.currentBlock.NewAlloca(arg.Type())
							c.memcpyInternal(alloca, arg.(*ir.InstLoad).Src)
							arg = alloca
						}

					}
					//param.xAttrs = slices.Delete(param.Attrs, 0, 1)
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
	case *types.StructType, *types.ArrayType, *types.IntType, *types.FloatType, *types.PointerType:

		var f *ir.Func
		if _, ok := c.funcs["GetSizeOf_"+t.Name()]; !ok {
			f = c.module.NewFunc("GetSizeOf_"+t.Name(), types.I64)
			block := f.NewBlock(entryBlock)
			alloca := block.NewAlloca(t)
			endPtr := block.NewGetElementPtr(t, alloca, constant.NewInt(types.I32, 1))
			ptrToInt := block.NewPtrToInt(alloca, types.I64)
			endPtrInt := block.NewPtrToInt(endPtr, types.I64)
			sizeInBytes := generateOperation(block, endPtrInt, ptrToInt, Sub)
			block.NewRet(sizeInBytes)
			c.funcs["GetSizeOf_"+t.Name()] = f
		} else {
			f = c.funcs["GetSizeOf_"+t.Name()]
		}

		return c.currentBlock.NewCall(f)
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
			var val = alloc
			if _, ok := alloc.(*ir.InstAlloca).ElemType.(*types.PointerType); ok {
				val = c.currentBlock.NewLoad(types.NewPointer(typ), alloc)
			}

			if c.isStorable(kv.Context()) {
				gep := c.currentBlock.NewGetElementPtr(typ, val, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, int64(fieldIndex)))
				c.currentBlock.NewStore(v, gep)
			} else {
				gep := c.currentBlock.NewGetElementPtr(typ, val, constant.NewInt(types.I32, 0), constant.NewInt(types.I32, int64(fieldIndex)))
				switch v := v.(type) {
				case *ir.InstLoad:
					c.memcpyInternal(gep, v.Src)
				default:
					c.memcpyInternal(gep, v)
				}
			}

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
			switch s1 := s.Src.(type) {
			case *ir.InstAlloca:
				elemType = s1.ElemType
				returnVal = s1
			case *ir.InstGetElementPtr:
				elemType = s1.ElemType
				returnVal = s
			case *ir.InstLoad:
				elemType = s1.ElemType
				returnVal = s1
			default:
				panic("unimplemented case")
			}
		case *ir.InstCall:
			elemType = s.Typ.(*types.PointerType).ElemType
			returnVal = s
		case *ir.InstAlloca:
			elemType = s.ElemType
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
	var elemType = src.Type().(*types.PointerType).ElemType
	if s, ok := src.(*ir.InstAlloca); ok {
		elemType = s.ElemType
		if e, ok := s.ElemType.(*types.PointerType); ok {
			elemType = e.ElemType
		}
	}
	sizeInBytes := c.callSizeOf(elemType)
	a1 := c.currentBlock.NewBitCast(dest, types.NewPointer(types.I8))
	a2 := c.currentBlock.NewBitCast(src, types.NewPointer(types.I8))
	return c.currentBlock.NewCall(f, a1, a2, sizeInBytes)
}

func (c *Compiler) mallocInternal(expr Expr) value.Value {
	callExpr := &CallExpr{
		Left: &IdentExpr{Name: "malloc_internal"},
		Args: []Expr{
			expr,
		},
		end:      Pos{},
		MethodOf: "",
		TypeCast: false,
	}

	return c.call(callExpr)
}

func (c *Compiler) getSizeOf(expr Expr) Expr {
	switch expr := expr.(type) {
	case *ObjectLitExpr:
		return &UnaryExpr{Op: Sizeof, Right: expr.Type, start: expr.start}
	case *IdentExpr:
		return &UnaryExpr{Op: Sizeof, Right: expr, start: expr.start}
	default:
		panic("TODO")
	}
}

func (c *Compiler) isStorable(ctx *VarAssignCtx) bool {

	if ctx.isPointer {
		return true
	}

	if ctx.ExpectedType() == "string" {
		return false
	}

	if ctx.isArray {
		return false
	}

	if ctx.isStruct {
		return false
	}
	if ctx.isSlice {
		return false
	}

	return true

}

func (c *Compiler) callCompareString(expr *BinaryExpr) value.Value {
	callResult := c.call(&CallExpr{
		Left: &IdentExpr{Name: "compare_string"},
		Args: []Expr{
			expr.Left,
			expr.Right,
		},
	})
	return generateOperation(c.currentBlock, callResult, constant.NewInt(types.I1, 1), expr.Op)
}

func (c *Compiler) callConcatString(expr *BinaryExpr) value.Value {
	return c.call(&CallExpr{
		Left: &IdentExpr{Name: "concat_strings"},
		Args: []Expr{
			expr.Left,
			expr.Right,
		},
	})
}

func (c *Compiler) callConcatCharAndString(expr *BinaryExpr) value.Value {
	return c.call(&CallExpr{
		Left: &IdentExpr{Name: "concat_char_and_string"},
		Args: []Expr{
			expr.Left, expr.Right,
		},
	})
}

func (c *Compiler) callConcatStringAndChar(expr *BinaryExpr) value.Value {
	return c.call(&CallExpr{
		Left: &IdentExpr{Name: "concat_string_and_char"},
		Args: []Expr{
			expr.Left, expr.Right,
		},
	})
}

func (c *Compiler) callConcatCharAndChar(expr *BinaryExpr) value.Value {
	return c.call(&CallExpr{
		Left: &IdentExpr{Name: "concat_char_and_char"},
		Args: []Expr{
			expr.Left, expr.Right,
		},
	})
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
				Size: &NumberExpr{Value: -1},
				end:  Pos{},
			}
		default:
			panic("not implemented ctx for array lit")
		}
		typ := c.resolveType(arrayType)
		if arrayType.Size.(*NumberExpr).Value == -1 {
			alloc := c.currentBlock.NewAlloca(c.types["slice"])
			c.currentBlock.NewCall(c.funcs["make_slice"], alloc)

			for _, elem := range expr.Elems {
				e := c.compileExpr(elem)
				c._callAppend(c.currentBlock.NewLoad(c.types["slice"], alloc), e)
			}

			return alloc
		} else {
			return c.newArray(expr, typ)
		}

	case *BinaryExpr:

		if expr.Ctx.IsRuntime {
			info := expr.Ctx.OpInfo
			if info.left == "string" && info.right == "string" {
				if info.op == Eq || info.op == Neq {
					return c.callCompareString(expr)
				} else if info.op == Add {
					return c.callConcatString(expr)
				} else {
					panic("Unsupported string operation")
				}
			}

			if info.left == "char" && info.op == Add && info.right == "string" {
				return c.callConcatCharAndString(expr)
			}

			if info.left == "string" && info.op == Add && info.right == "char" {
				return c.callConcatStringAndChar(expr)
			}

			if info.left == "char" && info.op == Add && info.right == "char" && expr.Ctx.parentCtx().(ExpectedTypeCtx).ExpectedType() == "string" {
				return c.callConcatCharAndChar(expr)
			}

			panic("Unimplemented runtime operation")

		}

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
		var size int
		if ok {
			Assert(ok, "Unimplemented number type")
			size = getBitSize(ctx.ExpectedType())
			if size == -1 {
				size = 32 // default
			}
		} else {
			size = 32
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
		strPtr.Align = 1
		strPtr.Immutable = true
		strPtr.UnnamedAddr = enum.UnnamedAddrUnnamedAddr
		var expectedType = "string"
		ctx, ok := expr.GetCtx().(ExpectedTypeCtx)
		if ok {
			expectedType = ctx.ExpectedType()
		}
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
		if ct, ok := expr.GetCtx().(*CallCtx); ok && ct.isCustomCall {
			return c.customCall(expr)
		} else {
			return c.call(expr)
		}
	case *IdentExpr:
		v := c.getAlloca(expr.Name)
		switch v := v.(type) {
		case *ir.InstAlloca:
			return c.currentBlock.NewLoad(v.ElemType, v)
		case *ir.Param:
			// TODO this is workaround
			if len(v.Attrs) > 0 {
				if _, ok := v.Attrs[0].(PassByValue); ok {
					alloca := c.currentBlock.NewAlloca(v.Typ.(*types.PointerType).ElemType)
					load := c.currentBlock.NewLoad(v.Typ.(*types.PointerType).ElemType, v)
					c.currentBlock.NewStore(load, alloca)

					return load
				}
			}
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
			return c.currentBlock.NewMul(constant.NewInt(right.Type().(*types.IntType), -1), right)
		case Band:
			switch e := expr.Right.(type) {
			case *ObjectLitExpr:
				size := c.getSizeOf(expr.Right)
				allocation := c.mallocInternal(size)
				typ := c.resolveType(e.Type)
				bitcast := c.currentBlock.NewBitCast(allocation, types.NewPointer(typ))
				var alloc = c.currentBlock.NewAlloca(types.NewPointer(typ))
				c.currentBlock.NewStore(bitcast, alloc)
				val := c.newObject(e, alloc)
				return c.currentBlock.NewLoad(types.NewPointer(typ), val)
			case *IdentExpr:
				return c.getAlloca(e.Name)
			case *SelectorExpr:
				gep := c.compileSelectorExpr(e)
				return gep.(*ir.InstLoad).Src
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
		case Not:
			right := c.compileExpr(expr.Right)
			return generateOperation(c.currentBlock, right, constant.NewInt(types.I1, 1), Xor)
		default:
			panic("Unreachable unary expression op = " + expr.Op.String())
		}
	case *AssignExpr:
		left := c.compileExpr(expr.Left)
		right := c.compileExpr(expr.Right)
		load, ok := left.(*ir.InstLoad)
		Assert(ok, "assign_expr left operand is not loaded")
		Assert(load.Src != nil, "assign_expr load.Src is nil")
		if c.isStorable(expr.Context()) {
			c.currentBlock.NewStore(right, load.Src)
		} else {
			var r = right
			if r2, ok := right.(*ir.InstLoad); ok {
				r = r2.Src
			}
			c.memcpyInternal(load.Src, r)
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
			} else if t.Name() == "slice" {

				sourceBaseType := expr.ctx.sourceBaseType
				if expr.ctx.isPointer {
					data := c.call(&CallExpr{
						Left: &IdentExpr{Name: "access_slice_datap"},
						Args: []Expr{
							expr.Left,
							expr.Index,
						},
					})
					alloc := c.currentBlock.NewAlloca(types.NewPointer(c.types[sourceBaseType]))
					c.currentBlock.NewStore(c.currentBlock.NewIntToPtr(data, types.NewPointer(c.types[sourceBaseType])), alloc)
					return c.currentBlock.NewLoad(types.NewPointer(c.types[sourceBaseType]), alloc)
				} else {
					data := c.call(&CallExpr{
						Left: &IdentExpr{Name: "access_slice_data"},
						Args: []Expr{
							expr.Left,
							expr.Index,
						},
					})
					itoptr := c.currentBlock.NewIntToPtr(data, types.NewPointer(c.types[sourceBaseType]))
					return c.currentBlock.NewLoad(c.types[sourceBaseType], itoptr)
				}
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
		return c.compileSelectorExpr(expr)
	default:
		panic("unreachable")
	}
}

func (c *Compiler) compileSelectorExpr(expr *SelectorExpr) value.Value {
	var selRes value.Value
	sel := c.compileExpr(expr.Selector)
	selRes = sel
	if sel, ok := sel.(*ir.InstAlloca); ok {
		selRes = c.currentBlock.NewLoad(sel.ElemType, sel)
	}
	selLoad, ok := selRes.(*ir.InstLoad)
	Assert(ok, "Selector invalid")
	if s, ok := selLoad.ElemType.(*types.PointerType); ok {
		selLoad = c.currentBlock.NewLoad(s.ElemType, selLoad)
	}
	zero := constant.NewInt(types.I32, 0)
	typName := selRes.Type().Name()
	if t, ok := selRes.Type().(*types.PointerType); ok {
		typName = t.ElemType.Name()
	}
	fieldIndex, err := c.resolveFieldIndex(expr.Ident.(*IdentExpr).Name, typName)
	AssertErr(err)
	idx := constant.NewInt(types.I32, int64(fieldIndex))
	gep := c.currentBlock.NewGetElementPtr(selLoad.ElemType, selLoad.Src, zero, idx)
	return c.currentBlock.NewLoad(c.getIndexedType(selLoad, fieldIndex), gep)
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
	Xor
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
	case Xor:
		return block.NewXor(value1, value2)
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
