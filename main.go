package main

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"log"
	"os"
	"os/exec"
	"strings"
)

func assert(cond bool, msg string) {
	if cond {
		return
	}
	panic(msg)
}

type Compiler struct {
	module    *ir.Module
	variables map[string]value.Value
	block     *ir.Block
	fun       *ir.Func
	funcs     map[string]*ir.Func
	types     map[string]types.Type
}

func (c *Compiler) initBuiltinTypes() {
	// we have already built-in types in llvm, so we don't have to re-define them
	// just add them to the dictionary so that make them accessible in the compiler
	if c.types == nil {
		c.types = make(map[string]types.Type)
	}

	c.types["int"] = types.I32
}

func (c *Compiler) initBuiltinFuncs() {
	module := c.module
	runtimePrintf := module.NewFunc("r_runtime_printf", types.I32,
		ir.NewParam("", types.NewPointer(types.I8)))
	runtimePrintf.Sig.Variadic = true
	runtimePrintf.Linkage = enum.LinkageExternal

	if c.funcs == nil {
		c.funcs = make(map[string]*ir.Func)
	}

	c.funcs["r_runtime_printf"] = runtimePrintf
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
	assert(c.types != nil, "un-initialized type map")

	panic("TODO")
}

func (c *Compiler) defineVar(expr *VarDefStmt) {
	assert(c.variables != nil, "un-initialized variable map")
	assert(c.block != nil, "block cannot be nil")
	block := c.block

	typ := c.resolveType(expr.Type)
	ptr := block.NewAlloca(typ)
	c.variables[expr.Name.Name] = ptr

	if expr.Init != nil {
		block.NewStore(c.expr(expr.Init), ptr)
	}

}

func (c *Compiler) newFunc(def *FuncDef) *ir.Func {
	assert(def.Name != nil, "type checker has to handle this")
	assert(def.Name.Name != "", "name can't be empty")
	name := def.Name.Name
	typ := c.resolveType(def.Type)

	params := make([]*ir.Param, 0)

	for _, def := range def.Params {
		typ2 := c.resolveType(def.Type)
		param := ir.NewParam(def.Name.Name, typ2)
		params = append(params, param)

		// TODO merge with defineVar function
		c.variables[def.Name.Name] = param
	}
	f := c.module.NewFunc(name, typ, params...)
	c.funcs[name] = f
	c.fun = f
	block := f.NewBlock(entryBlock)
	c.block = block

	if stmt, ok := def.Body.(*BlockStmt); ok {
		for _, innerStmt := range stmt.Stmts {
			switch innerStmt := innerStmt.(type) {
			case *ReturnStmt:
				block.NewRet(c.expr(innerStmt.Value))
			case *VarDefStmt:
				c.defineVar(innerStmt)
			}
		}
	}

	return f
}

func (c *Compiler) getVariable(name string) value.Value {
	if v, ok := c.variables[name]; ok {
		return v
	}
	panic("no such variable: " + name)
}

func (c *Compiler) expr(expr Expr) value.Value {
	switch expr := expr.(type) {
	case *BinaryExpr:
		left := c.expr(expr.Left)
		right := c.expr(expr.Right)
		return generateOperation(c.block, left, right, expr.Op)
	case *NumberExpr:
		return getI32Constant(expr.Value)
	case *IdentExpr:
		v := c.getVariable(expr.Name)
		switch v := v.(type) {
		case *ir.Param:
			return v
		case *ir.InstAlloca:
			return c.block.NewLoad(types.I32, v)
		default:
			panic("TODO:")
		}
	default:
		panic("unreachable")
	}
}

const entryBlock string = "entry"

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
)

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
	}

	panic("unreachable")

}

type Expr interface {
	IsExpr()
}

type NumberExpr struct {
	Value int64
}

func (e *NumberExpr) IsExpr() {}

type BinaryExpr struct {
	Left  Expr
	Right Expr
	Op    Operation
}

func (e *BinaryExpr) IsExpr() {}

type CallExpr struct {
	Func   Expr
	Params []Expr
}

func isWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isDigit(c rune) bool {
	return '0' <= c && c <= '9'
}

func parseBinaryExpression() *BinaryExpr {
	return &BinaryExpr{
		Left: &BinaryExpr{
			Left: &BinaryExpr{
				Left:  &NumberExpr{Value: 36},
				Right: &NumberExpr{Value: 5},
				Op:    Div,
			},
			Right: &NumberExpr{Value: 3},
			Op:    Div,
		},
		Right: &NumberExpr{Value: 1},
		Op:    Div,
	}
}

type IdentExpr struct {
	Name string
}

func (e *IdentExpr) IsExpr() {}

type FuncDef struct {
	Name   *IdentExpr
	Type   Expr
	Params []*ParamExpr
	Body   Stmt
}

type ParamExpr struct {
	Name *IdentExpr
	Type Expr
}

type BlockStmt struct {
	Stmts []Stmt
}

func (s *BlockStmt) IsStmt() {}

func (e *ParamExpr) IsExpr() {}

type Stmt interface {
	IsStmt()
}
type ReturnStmt struct {
	Value Expr
}

func (e *ReturnStmt) IsStmt() {}

type VarDefStmt struct {
	Name *IdentExpr
	Type Expr
	Init Expr
}

func (e *VarDefStmt) IsStmt() {}

func generateFuncDefs() []*FuncDef {
	funcDefs := make([]*FuncDef, 0)

	def := &FuncDef{}

	def.Name = &IdentExpr{Name: "ExampleFunc"}
	def.Type = &IdentExpr{Name: "int"}
	def.Params = []*ParamExpr{
		{Name: &IdentExpr{Name: "a"}, Type: &IdentExpr{Name: "int"}},
		{Name: &IdentExpr{Name: "b"}, Type: &IdentExpr{Name: "int"}},
	}

	def.Body = &BlockStmt{Stmts: []Stmt{}}
	def.Body.(*BlockStmt).Stmts = append(def.Body.(*BlockStmt).Stmts, &VarDefStmt{Name: &IdentExpr{Name: "c"}, Type: &IdentExpr{Name: "int"}, Init: &NumberExpr{Value: 992}})
	def.Body.(*BlockStmt).Stmts = append(def.Body.(*BlockStmt).Stmts, &ReturnStmt{Value: &BinaryExpr{
		Left:  &IdentExpr{Name: "a"},
		Right: &IdentExpr{Name: "c"},
		Op:    Add,
	}})

	funcDefs = append(funcDefs, def)
	return funcDefs
}

func generateCallExprs() []*CallExpr {
	callExprs := make([]*CallExpr, 0)

	callExprs = append(callExprs, &CallExpr{
		Func: &IdentExpr{Name: "ExampleFunc"},
		Params: []Expr{
			&BinaryExpr{
				Left:  &NumberExpr{Value: 0},
				Right: &NumberExpr{Value: 1},
				Op:    Add,
			},
			&NumberExpr{Value: 1},
		},
	})

	return callExprs
}

func (c *Compiler) init() {
	assert(c.module != nil, "module not initialized")
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

func main() {

	_ = os.Setenv("DEBUG", "1")

	compiler := &Compiler{}

	// read expr from file
	file, err := os.ReadFile("./input.file")
	_ = file
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	outputPath := "./plus.ll"
	newFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}

	compiler.module = &ir.Module{}
	compiler.init()

	mainFunc := compiler.module.NewFunc("main", types.I32)
	block := mainFunc.NewBlock(entryBlock)

	g := compiler.module.NewGlobalDef("str", constant.NewCharArrayFromString("hello %d\n"))

	funcDefs := generateFuncDefs()
	for _, def := range funcDefs {
		compiler.newFunc(def)
	}

	compiler.block = block
	compiler.fun = mainFunc
	for _, callExpr := range generateCallExprs() {
		passParams := make([]value.Value, 0)
		for _, p := range callExpr.Params {
			v := compiler.expr(p)
			passParams = append(passParams, v)
		}

		callOp := block.NewCall(compiler.getFunc(callExpr.Func.(*IdentExpr).Name), passParams...)
		block.NewCall(compiler.getFunc("r_runtime_printf"), g, callOp)
	}

	alloc := block.NewAlloca(types.I32)
	block.NewStore(getI32Constant(111), alloc)
	loaded := block.NewLoad(types.I32, alloc)
	block.NewCall(compiler.getFunc("r_runtime_printf"), g, loaded)
	block.NewRet(loaded)
	_, err = compiler.module.WriteTo(newFile)
	if err != nil {
		panic(err)
	}
	clangCompile(outputPath, "./runtime/runtime.c", true, true)

}

func clangCompile(path, runtimePath string, runBinary bool, outputForwarding bool) {

	var outputPath = path[:strings.LastIndex(path, ".")] + ".o"

	// clang ${path} -o ${path}.o

	runtimeOutputPath := runtimePath[:strings.LastIndex(runtimePath, ".")] + ".ll"
	runtimeCompile := exec.Command("clang", runtimePath, "-S", "-emit-llvm", "-o", runtimeOutputPath)
	if outputForwarding {
		runtimeCompile.Stdout = os.Stdout
		runtimeCompile.Stderr = os.Stdout
	}
	if err := runtimeCompile.Run(); err != nil {
		log.Fatalf("failed to run command for clang: %v", err)
	}

	command := exec.Command("clang", path, runtimeOutputPath, "-o", outputPath)
	if outputForwarding {
		command.Stdout = os.Stdout
		command.Stderr = os.Stdout
	}
	err := command.Run()

	if err != nil {
		log.Fatalf("failed to run command for clang: %v", err)
	}

	if runBinary {
		runBinaryCmd := exec.Command(outputPath)
		runBinaryCmd.Stdin = os.Stdin
		runBinaryCmd.Stdout = os.Stdout
		runBinaryCmd.Stderr = os.Stderr

		if err := runBinaryCmd.Run(); err != nil {
		}
	}

}
