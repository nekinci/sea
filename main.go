package main

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"
)

const entryBlock string = "entry"

func assert(cond bool, msg string) {
	if cond {
		return
	}
	panic(msg)
}

type Compiler struct {
	module        *ir.Module
	variables     map[string]value.Value
	currentBlock  *ir.Block
	breakBlock    *ir.Block
	continueBlock *ir.Block

	currentFunc *ir.Func
	funcs       map[string]*ir.Func
	types       map[string]types.Type
	pkg         *Package
}

func (c *Compiler) compile() {
	assert(c.pkg != nil, "pkg is nil")
	assert(c.module != nil, "module is nil")
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

func (c *Compiler) compileVarDef(stmt *VarDefStmt) {
	assert(c.variables != nil, "un-initialized variable map")
	assert(c.currentBlock != nil, "currentBlock cannot be nil")
	block := c.currentBlock

	typ := c.resolveType(stmt.Type)
	ptr := block.NewAlloca(typ)
	c.variables[stmt.Name.Name] = ptr

	if stmt.Init != nil {
		block.NewStore(c.compileExpr(stmt.Init), ptr)
	}
}

func (c *Compiler) compileFunc(def *FuncDefStmt) *ir.Func {
	assert(def.Name != nil, "type checker has to handle this")
	assert(def.Name.Name != "", "name can't be empty")
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
	funcBlock := c.currentFunc.NewBlock("funcBlock")
	initBlock := c.currentFunc.NewBlock("initBlock")
	forBlock := c.currentFunc.NewBlock("forBlock")
	condBlock := c.currentFunc.NewBlock("condBlock")
	stepBlock := c.currentFunc.NewBlock("StepBlock")

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

func (c *Compiler) compileExpr(expr Expr) value.Value {
	switch expr := expr.(type) {
	case *BinaryExpr:
		left := c.compileExpr(expr.Left)
		right := c.compileExpr(expr.Right)
		return generateOperation(c.currentBlock, left, right, expr.Op)
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
		switch v := v.(type) {
		case *ir.Param:
			return v
		case *ir.InstAlloca:
			return c.currentBlock.NewLoad(v.ElemType, v)
		default:
			panic("TODO:")
		}

	case *ParenExpr:
		return c.compileExpr(expr.Expr)
	case *BoolExpr:
		return constant.NewBool(expr.Value)
	case *UnaryExpr:
		right := c.compileExpr(expr.Right)
		switch expr.Op {
		case Sub:
			return c.currentBlock.NewMul(constant.NewInt(types.I32, -1), right)
		case Band:
			val := c.compileExpr(expr.Right)
			return val.(*ir.InstLoad).Src
		default:
			panic("Unreachable unary expression op = " + expr.Op.String())
		}
	case *AssignExpr:
		right := c.compileExpr(expr.Expr)
		c.currentBlock.NewStore(right, c.getAlloca(expr.Name.Name))
		return c.compileExpr(expr.Name)
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

type Package struct {
	Name  string
	Stmts []Stmt
}

type Expr interface {
	Node
	IsExpr()
}

type NumberExpr struct {
	Value int64
}

type ParenExpr struct {
	Expr Expr
}

func (p *ParenExpr) IsExpr() {}

type StringExpr struct {
	Value    string
	Unquoted string
}

func (s *StringExpr) IsExpr() {}

func (e *NumberExpr) IsExpr() {}

type BoolExpr struct {
	Value bool // 1 true 0 false
}

func (b *BoolExpr) IsExpr() {}

type UnaryExpr struct {
	Op    Operation
	Right Expr
}

type AssignExpr struct {
	Name *IdentExpr
	Expr Expr
}

func (a *AssignExpr) IsExpr() {}

func (u *UnaryExpr) IsExpr() {}

type BinaryExpr struct {
	Left  Expr
	Right Expr
	Op    Operation
}

func (e *BinaryExpr) IsExpr() {}

type CallExpr struct {
	Name *IdentExpr
	Args []Expr
}

func (c *CallExpr) IsExpr() {}

func isWhitespace(c uint8) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isDigit(c uint8) bool {
	return '0' <= c && c <= '9'
}

type IdentExpr struct {
	Name string
}

func (e *IdentExpr) IsExpr() {}

type FuncDefStmt struct {
	Name       *IdentExpr
	Type       Expr
	Params     []*ParamExpr
	Body       *BlockStmt
	IsExternal bool
}

func (f *FuncDefStmt) IsStmt() {}
func (f *FuncDefStmt) IsDef()  {}

type Node interface{}

type ParamExpr struct {
	Name *IdentExpr
	Type Expr
}

type BlockStmt struct {
	Stmts []Stmt
}

type ImplStmt struct {
	Stmts   []Stmt
	Implies []*IdentExpr
}

func (e *ImplStmt) IsStmt() {}

type DefStmt interface {
	IsDef()
}

func (s *BlockStmt) IsStmt() {}

func (e *ParamExpr) IsExpr() {}

type Stmt interface {
	Node
	IsStmt()
}

type IfStmt struct {
	Cond Expr
	Then Stmt // what about single line un-blocked compileStmt?
	Else Stmt
}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Step Expr
	Body Stmt
}

func (f *ForStmt) IsStmt() {}

func (e *IfStmt) IsStmt() {}

type ReturnStmt struct {
	Value Expr
}

// BreakStmt TODO support labels
type BreakStmt struct {
}

func (b *BreakStmt) IsStmt() {}

type ContinueStmt struct{}

func (c *ContinueStmt) IsStmt() {}

func (e *ReturnStmt) IsStmt() {}

type VarDefStmt struct {
	Name  *IdentExpr
	Type  Expr
	IsPtr bool
	Init  Expr
}

func (e *VarDefStmt) IsDef() {}

func (e *VarDefStmt) IsStmt() {}

type StructDefStmt struct {
	Name   *IdentExpr
	Fields []*Field
}

func (s *StructDefStmt) IsDef() {}

func (s *StructDefStmt) IsStmt() {}

type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) IsStmt() {}

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

	// read compileExpr from file
	file, err := os.ReadFile("./input.file")
	parser := NewParser("./input.file")
	//parser.printTokens()
	pckg := parser.parse()
	writer := NewASTWriter(io.Discard, pckg)
	writer.start()

	_ = file
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	outputPath := "./plus.ll"
	newFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}

	compiler.module = ir.NewModule()
	compiler.init()
	compiler.pkg = pckg
	compiler.compile()
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
			os.Exit(runBinaryCmd.ProcessState.ExitCode())
		}
	}

}

/// Lexer and Parser

func (t Token) String() string {
	return string(t)
}

type Lexer struct {
	filename string
	input    string
	pos      int
	start    int
	inputLen int
	curTok   Token
	curVal   string
	comments bool
}

func (l *Lexer) operatorToken() Token {
	ch := l.input[l.pos]
	l.pos = l.pos + 1

	switch ch {
	case '+':
		if l.input[l.pos] == '+' {
			// TODO
		}

		if l.input[l.pos] == '=' {
			// TODO
		}

		return TokPlus
	case '-':
		if l.input[l.pos] == '-' {
			// TODO
		}

		if l.input[l.pos] == '=' {
			// TODO
		}

		return TokMinus
	case '*':
		if l.input[l.pos] == '=' {
		}
		return TokMultiply
	case '/':
		if l.input[l.pos] == '=' {
		}
		return TokDivision
	case '%':
		if l.input[l.pos] == '=' {
		}
		return TokMod
	}

	panic("unreachable")

}

func (l *Lexer) printTokens() {
	next, s := l.next()
	for next != EOF {
		fmt.Printf("Token[%s - %s]\n", next, s)
		next, s = l.next()
	}

}

func (l *Lexer) backup(len int) {
	l.pos = l.pos - len
	assert(l.pos >= 0 && l.pos < l.inputLen, "lexer position out of bounds")
}

func (l *Lexer) nextAndBackup() (Token, string) {
	token, v := l.next()
	l.backup(len(v))
	return token, v
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.input[l.pos]) {
		l.pos = l.pos + 1
	}
}

func (l *Lexer) skipLineComment() Token {
	for l.input[l.pos] != '\n' {
		l.pos = l.pos + 1
	}
	l.pos = l.pos + 1
	return TokSingleComment
}

func (l *Lexer) skipBlockComment() Token {
	for l.input[l.pos] != '*' {
		l.pos = l.pos + 1
	}
	l.pos = l.pos + 1
	for l.input[l.pos] != '/' {
		l.pos = l.pos + 1
	}
	l.pos = l.pos + 1
	return TokMultiComment
}

func (l *Lexer) tok() Token {
	if l.pos >= len(l.input) {
		return EOF
	}

	l.start = l.pos
	c := l.input[l.pos]
	switch {
	case c == '#':
		l.skipLineComment()
		return l.tok()
	case isWhitespace(c):
		l.skipWhitespace()
		return l.tok()
	case isDigit(c):
		// TODO handle floating point numbers, negative numbers or e signed integers
		for c >= '0' && c <= '9' {
			l.pos++
			c = l.input[l.pos]
		}
		return TokNumber
	case c == '+' || c == '-' || c == '*' || c == '/' || c == '%':
		if c == '/' && l.inputLen > l.pos+1 && l.input[l.pos+1] == '*' {
			l.pos += 2
			for l.pos+1 < l.inputLen && !(l.input[l.pos] == '*' && l.input[l.pos+1] == '/') {
				l.pos++
			}
			l.pos += 2
			return l.tok()
		}
		return l.operatorToken()
	case c == '{':
		l.pos++
		return TokLBrace
	case c == '}':
		l.pos++
		return TokRBrace
	case c == '(':
		l.pos++
		return TokLParen
	case c == ')':
		l.pos++
		return TokRParen
	case c == ',':
		l.pos++
		return TokComma
	case c == ';':
		l.pos++
		return TokSemicolon
	case unicode.IsLetter(rune(c)):

		for l.pos < l.inputLen && (unicode.IsDigit(rune(l.input[l.pos])) || unicode.IsLetter(rune(l.input[l.pos])) || l.input[l.pos] == '_') {
			l.pos += 1
		}

		identifier := l.value()
		switch identifier {
		case "var":
			return TokVar
		case "fun":
			return TokFun
		case "return":
			return TokReturn
		case "true":
			return TokTrue
		case "false":
			return TokFalse
		case "if":
			return TokIf
		case "else":
			return TokElse
		case "for":
			return TokFor
		case "break":
			return TokBreak
		case "continue":
			return TokContinue
		case "extern":
			return TokExtern
		case "struct":
			return TokStruct
		case "impl":
			return TokImpl
		}

		return TokIdentifier
	case isDigit(c):
		// TODO . , e handle
		for l.pos < l.inputLen && isDigit(l.input[l.pos]) {
			l.pos += 1
		}
		return TokNumber
	case c == '"':
		first := true
		for l.pos < l.inputLen && (l.input[l.pos] != '"' || first) {
			l.pos += 1
			first = false
		}
		l.pos++
		return TokString
	case c == '=':
		l.pos += 1
		if l.input[l.pos] == '=' {
			l.pos += 1
			return TokEqual
		}

		return TokAssign
	case c == '!':
		l.pos++
		if l.input[l.pos] == '=' {
			l.pos += 1
			return TokNEqual
		}
		return TokNot
	case c == '&':
		l.pos++
		if l.input[l.pos] == '&' {
			l.pos += 1
			return TokAnd
		}
		return TokBAnd
	case c == '|':
		l.pos++
		if l.input[l.pos] == '|' {
			l.pos += 1
			return TokOr
		}
		return TokBOr
	case c == '^':
		l.pos++
		return TokXor
	case c == '<':
		l.pos++
		if l.input[l.pos] == '=' {
			l.pos += 1
			return TokLte
		}

		if l.input[l.pos] == '<' {
			l.pos++
			return TokLShift
		}

		return TokLt

	case c == '>':
		l.pos++
		if l.input[l.pos] == '=' {
			l.pos += 1
			return TokGte
		}

		if l.input[l.pos] == '>' {
			l.pos++
			return TokRShift
		}

		return TokGt

	}

	panic("unreachable: " + string(rune(c)))
}

func (l *Lexer) next() (Token, string) {

	tok := l.tok()
	l.curTok = tok
	l.curVal = l.value()

	return tok, l.curVal

}

func (l *Lexer) value() string {
	return l.input[l.start:l.pos]
}

type Parser struct {
	*Lexer
	module *Package
	errors []string
}

func NewParser(path string) *Parser {
	parser := &Parser{}
	assert(path != "", "path cannot be empty")
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	lexer := &Lexer{
		filename: path,
		input:    string(file),
		pos:      0,
		start:    0,
		inputLen: len(file),
	}

	parser.Lexer = lexer
	return parser
}

func (p *Parser) parse() *Package {
	pckg := &Package{
		Name:  "Main",
		Stmts: []Stmt{},
	}
	p.module = pckg

	p.next()
	for p.curTok != EOF {
		pckg.Stmts = append(pckg.Stmts, p.parseStmt())
	}

	return pckg

}

func (p *Parser) parseVar() *VarDefStmt {
	p.expect(TokVar)
	_, typ := p.expect(TokIdentifier)
	_, name := p.expect(TokIdentifier)

	isPtr := false
	if p.curTok == TokMultiply {
		p.expect(TokMultiply)
		isPtr = true
	}

	varDefStmt := &VarDefStmt{
		Name:  &IdentExpr{Name: name},
		Type:  &IdentExpr{Name: typ},
		Init:  nil,
		IsPtr: isPtr,
	}

	if p.curTok == TokAssign {
		p.expect(TokAssign)
		varDefStmt.Init = p.parseExpr()
	}
	return varDefStmt
}

func (p *Parser) parseIf() *IfStmt {
	p.expect(TokIf)
	ifStmt := &IfStmt{}
	ifStmt.Cond = p.parseExpr()
	ifStmt.Then = p.parseStmt()
	if p.curTok == TokElse {
		p.expect(TokElse)
		ifStmt.Else = p.parseStmt()
	}
	return ifStmt
}

type Field struct {
	Name *IdentExpr
	Type *IdentExpr
}

func (f *Field) IsExpr() {}

func (p *Parser) parseField() *Field {
	_, typ := p.expect(TokIdentifier)
	_, name := p.expect(TokIdentifier)
	return &Field{
		Name: &IdentExpr{Name: name},
		Type: &IdentExpr{Name: typ},
	}
}

func (p *Parser) parseStruct() *StructDefStmt {
	p.expect(TokStruct)
	_, name := p.expect(TokIdentifier)
	p.expect(TokLBrace)
	fields := make([]*Field, 0)

	for p.curTok != TokRBrace {
		fields = append(fields, p.parseField())
	}

	p.expect(TokRBrace)

	return &StructDefStmt{
		Name:   &IdentExpr{Name: name},
		Fields: fields,
	}
}

func (p *Parser) parseFor() *ForStmt {
	p.expect(TokFor)
	forStmt := &ForStmt{}

	// TODO add support multiple variable stmts
	if p.curTok == TokVar {
		varDefStmt := p.parseVar()
		forStmt.Init = varDefStmt
		p.expect(TokSemicolon)
	}

	forStmt.Cond = p.parseExpr()
	if p.curTok == TokSemicolon {
		p.expect(TokSemicolon)
		forStmt.Step = p.parseExpr()
	}

	forStmt.Body = p.parseStmt()

	return forStmt
}

func (p *Parser) parseStmt() Stmt {

	switch p.curTok {
	case TokExtern:
		if p.mayNextBe(TokFun) {
			return p.parseFunc()
		} else if p.mayNextBe(TokVar) {
			return p.parseVar()
		}
		p.assume(false, "extern cannot be combined except fun and var")
	case TokFun:
		return p.parseFunc()
	case TokLBrace:
		return p.parseBlock()
	case TokBreak:
		p.expect(TokBreak)
		// TODO handle labels
		return &BreakStmt{}
	case TokContinue:
		p.expect(TokContinue)
		return &ContinueStmt{}
	case TokReturn:
		p.expect(TokReturn)
		returnStmt := &ReturnStmt{}
		returnStmt.Value = p.parseExpr()
		return returnStmt
	case TokVar:
		return p.parseVar()
	case TokIf:
		return p.parseIf()
	case TokFor:
		return p.parseFor()
	case TokStruct:
		return p.parseStruct()
	case TokImpl:
		return p.parseImpl()
	}

	return p.parseExprStmt()

}

func (p *Parser) parseImpl() *ImplStmt {
	panic("TODO")
}

func (p *Parser) parseBlock() *BlockStmt {
	p.expect(TokLBrace)
	blockStmt := &BlockStmt{Stmts: make([]Stmt, 0)}
	for {
		stmt := p.parseStmt()
		blockStmt.Stmts = append(blockStmt.Stmts, stmt)
		if p.curTok == TokRBrace {
			p.expect(TokRBrace)
			break
		}
	}

	return blockStmt
}

func (p *Parser) parseExprList() []Expr {
	exprs := make([]Expr, 0)

	for {
		expr := p.parseExpr()
		exprs = append(exprs, expr)
		if p.curTok == TokComma {
			p.expect(TokComma)
		} else {
			break
		}
	}

	return exprs
}

func (p *Parser) binaryPrecedence(tok Token) int {

	// If it is binary expression then minimum binaryPrecedence is 1; otherwise return 0 that indicates it is not binary operator
	// Operation precedence
	// 1. Multiply, Division, Lshift, Rshift, Mod
	// 2. Plus, Minus
	// 3. gt, lt, gte, lte,neq, eq
	// 4. and, band
	// 5. or, bor, xor
	// And also unary operators are more precise

	switch tok {
	case TokMultiply, TokDivision, TokLShift, TokRShift, TokMod:
		return 5
	case TokPlus, TokMinus:
		return 4
	case TokGt, TokGte, TokLt, TokLte, TokNEqual, TokEqual:
		return 3
	case TokAnd, TokBAnd:
		return 2
	case TokOr, TokBOr, TokXor:
		return 1
	}

	return 0
}

func (p *Parser) unaryPrecedence(tok Token) int {
	switch tok {
	case TokMultiply, TokBAnd:
		return 7
	case TokMinus, TokPlus, TokNot:
		return 6
	default:
		return 0
	}
}

func (p *Parser) parseBinaryExpr(parentPrecedence int) Expr {
	// 5 + 4 * 3
	// BinaryExpr[5 + BinaryExpr[4*3]]

	// 5 * 4 + 3
	// BinaryExpr[BinaryExpr[5*4] + 3]

	// 5*4*3
	// BinaryExpr[BinaryExpr[5*4] * 3]

	// -5+3

	var left Expr
	unaryPrecedence := p.unaryPrecedence(p.curTok)
	if unaryPrecedence != 0 && unaryPrecedence >= parentPrecedence {
		tok, _ := p.expectAnyOf(TokMinus, TokNot, TokPlus, TokBAnd)
		left = &UnaryExpr{
			Op:    p.op(tok),
			Right: p.parseBinaryExpr(unaryPrecedence),
		}
	} else {
		left = p.parseSimpleExpr()
	}

	precedence := p.binaryPrecedence(p.curTok)

	for precedence > 0 && precedence > parentPrecedence {
		var op = p.op(p.curTok)
		p.next()
		left = &BinaryExpr{
			Left:  left,
			Right: p.parseBinaryExpr(precedence),
			Op:    op,
		}
		precedence = p.binaryPrecedence(p.curTok)

	}

	return left
}

func (p *Parser) op(tok Token) Operation {
	switch tok {
	case TokPlus:
		return Add
	case TokMinus:
		return Sub
	case TokDivision:
		return Div
	case TokMultiply:
		return Mul
	case TokEqual:
		return Eq
	case TokGt:
		return Gt
	case TokGte:
		return Gte
	case TokLt:
		return Lt
	case TokLte:
		return Lte
	case TokNEqual:
		return Neq
	case TokAnd:
		return And
	case TokOr:
		return Or
	case TokBAnd:
		return Band
	default:
		panic("unreachable op=" + string(tok))
	}
}

func (p *Parser) parseSimpleExpr() Expr {
	switch p.curTok {
	case TokString:
		_, str := p.expect(TokString)
		unquotedStr, err := strconv.Unquote(str)
		assert(err == nil, "unquoted string expected")
		// implicitly add end of string to unquoted string
		unquotedStr += "\x00"
		return &StringExpr{Value: str, Unquoted: unquotedStr}
		return &StringExpr{Value: str, Unquoted: unquotedStr}
	case TokNumber:
		_, val := p.expect(TokNumber)
		v, err := strconv.ParseInt(val, 10, 64)
		assert(err == nil, fmt.Sprintf("Invalid number: %v", err))
		return &NumberExpr{Value: v}
	case TokTrue, TokFalse:
		tok, _ := p.expectAnyOf(TokTrue, TokFalse)
		return &BoolExpr{Value: tok == TokTrue}

	case TokLParen:
		p.expect(TokLParen)
		expr := p.parseExpr()
		p.expect(TokRParen)
		return &ParenExpr{Expr: expr}
	case TokIdentifier:
		return p.parseIdentExpr()
	default:
		panic("unreachable tok -> " + string(p.curTok))

	}
}

func (p *Parser) parseIdentExpr() Expr {
	_, identifier := p.expect(TokIdentifier)
	if p.curTok == TokLParen {
		callExpr := &CallExpr{}
		callExpr.Name = &IdentExpr{Name: identifier}
		p.expect(TokLParen)
		callExpr.Args = p.parseExprList()
		p.expect(TokRParen)

		return callExpr
	}

	if p.curTok == TokAssign {
		p.expect(TokAssign)
		assignExpr := &AssignExpr{&IdentExpr{Name: identifier}, p.parseExpr()}
		return assignExpr
	}

	return &IdentExpr{Name: identifier}
}

func (p *Parser) parseExpr() Expr {

	switch p.curTok {
	default:
		return p.parseBinaryExpr(0)
	}

}

func (p *Parser) parseExprStmt() *ExprStmt {
	return &ExprStmt{Expr: p.parseExpr()}
}

func (p *Parser) parseParams() []*ParamExpr {
	p.expect(TokLParen)
	params := make([]*ParamExpr, 0)
	for p.curTok != TokRParen {
		_, typeVal := p.expect(TokIdentifier)
		_, identVal := p.expect(TokIdentifier)
		param := &ParamExpr{
			Name: &IdentExpr{Name: identVal},
			Type: &IdentExpr{Name: typeVal},
		}

		params = append(params, param)

		if p.curTok == TokComma {
			p.expect(TokComma)
		}

	}

	p.expect(TokRParen)
	return params
}

func (p *Parser) assume(cond bool, msg string) {
	if !cond {
		p.errors = append(p.errors, msg)
	}
}

func (p *Parser) expect(tok Token) (Token, string) {
	defer p.next()
	if p.curTok == tok {
		return p.curTok, p.curVal
	}

	p.assume(tok == p.curTok, fmt.Sprintf("Expected %s, got %s", tok, p.curTok))
	return TokUnexpected, p.curVal
}

func (p *Parser) expectAnyOf(tokens ...Token) (Token, string) {
	defer p.next()
	for _, tok := range tokens {
		if p.curTok == tok {
			return p.curTok, p.curVal
		}
	}

	p.assume(false, fmt.Sprintf("Expected %s, got %s", tokens, p.curTok))
	return TokUnexpected, p.curVal
}

func (p *Parser) mayNextBe(tok Token) bool {
	t, v := p.curTok, p.curVal
	b, _ := p.nextAndBackup()
	p.curTok, p.curVal = t, v
	return b == tok
}

func (p *Parser) parseFunc() *FuncDefStmt {

	var isExternal bool
	if p.curTok == TokExtern {
		isExternal = true
		p.expect(TokExtern)
	}

	p.expect(TokFun)

	funcDef := &FuncDefStmt{
		Params:     make([]*ParamExpr, 0),
		IsExternal: isExternal,
	}
	_, identifier := p.expect(TokIdentifier)
	funcDef.Type = &IdentExpr{Name: identifier}
	_, identifier = p.expect(TokIdentifier)
	funcDef.Name = &IdentExpr{Name: identifier}
	funcDef.Params = p.parseParams()
	if !isExternal {
		funcDef.Body = p.parseBlock()
	}
	return funcDef
}

/*
	Writer
*/

type ASTWriter struct {
	io.Writer
	pkg    *Package
	indent int
}

func NewASTWriter(w io.Writer, p *Package) *ASTWriter {
	return &ASTWriter{w, p, 0}
}

func (w *ASTWriter) Dump() {
	w.start()
}

func (w *ASTWriter) start() {
	assert(w.pkg != nil, "AST is nil")
	_, err := fmt.Fprintf(w.Writer, "Package[%s]\n", w.pkg.Name)
	if err != nil {
		return
	}

	for _, stmt := range w.pkg.Stmts {
		w.indentStmt(w.writeStmt, stmt)
	}
}

func (w *ASTWriter) writeIndent() {
	for i := 0; i < w.indent; i++ {
		_, err := fmt.Fprintf(w.Writer, " ")
		if err != nil {
			return
		}
	}
}
func (w *ASTWriter) writeStmt(stmt Stmt) {
	w.writeIndent()
	switch s := stmt.(type) {
	case *FuncDefStmt:
		_, _ = fmt.Fprintf(w.Writer, "Function[Name: %s, Returns: %s]\n", s.Name.Name, s.Type.(*IdentExpr).Name)
		for _, param := range s.Params {
			w.indentExpr(w.writeExpr, param)
		}
		if !s.IsExternal {
			for _, stmt := range s.Body.Stmts {
				w.indentStmt(w.writeStmt, stmt)
			}
		}
	case *ExprStmt:
		w.indentExpr(w.writeExpr, s.Expr)
	}
}

func (w *ASTWriter) writeExpr(expr Expr) {
	w.writeIndent()
	switch e := expr.(type) {
	case *CallExpr:
		_, _ = fmt.Fprintf(w.Writer, "Call[%s]\n", e.Name.Name)
		for _, param := range e.Args {
			w.indentExpr(w.writeExpr, param)
		}
	case *ParamExpr:
		_, _ = fmt.Fprintf(w.Writer, "Param[Name: %s, Type: %s]\n", e.Name.Name, e.Type.(*IdentExpr).Name)
	case *StringExpr:
		_, _ = fmt.Fprintf(w.Writer, "String[%s]\n", e.Value)
	case *NumberExpr:
		_, _ = fmt.Fprintf(w.Writer, "Number[%d]\n", e.Value)
	case *IdentExpr:
		_, _ = fmt.Fprintf(w.Writer, "Ident[%s]\n", e.Name)
	}
}

func (w *ASTWriter) indentStmt(callback func(stmt Stmt), stmt Stmt) {
	w.indent += 2
	callback(stmt)
	w.indent -= 2
}

func (w *ASTWriter) indentExpr(callback func(expr Expr), expr Expr) {
	w.indent += 2
	callback(expr)
	w.indent -= 2
}
