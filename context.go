package main

import "github.com/llir/llvm/ir"

type Ctx interface {
	parentCtx() Ctx
}

type ExpectedTypeCtx interface {
	Ctx
	ExpectedType() string
}

type FuncCtx struct {
	parent        Ctx
	sym           *FuncDef
	isProblematic bool
	expectedType  string
	returnTypeSym *TypeDef
	returnsStruct bool
	returnParam   *ir.Param
}

func (f *FuncCtx) ExpectedType() string {
	return f.expectedType
}

func (f *FuncCtx) parentCtx() Ctx {
	return f.parent
}

type VarAssignCtx struct {
	parent        Ctx
	expectedType  string
	arraySize     int // optional for arrays
	isStruct      bool
	typSym        *TypeDef
	isPointer     bool
	isArray       bool
	isSlice       bool
	extractedType string
	isGlobal      bool
	isInitiated   bool

	// compiler side fields
	alloca *ir.InstAlloca
}

func (v *VarAssignCtx) ExpectedType() string {
	return v.expectedType
}

func (v *VarAssignCtx) parentCtx() Ctx {
	return v.parent
}

type CallCtx struct {
	parent            Ctx
	isCustomCall      bool
	returnsStruct     bool
	typeCastParamType string
}

func (c *CallCtx) parentCtx() Ctx {
	return c.parent
}

type IndexCtx struct {
	parent         Ctx
	expectedType   string
	sourceBaseType string
	isPointer      bool
}

func (i *IndexCtx) parentCtx() Ctx {
	return i.parent
}

func (i *IndexCtx) ExpectedType() string {
	return i.expectedType
}

type BinaryExprCtx struct {
	parent       Ctx
	IsRuntime    bool
	OpInfo       operationInfo
	expectedType string
	ResultType   string
}

func (b *BinaryExprCtx) ExpectedType() string {
	return b.expectedType
}

func (b *BinaryExprCtx) parentCtx() Ctx {
	return b.parent
}

type ForCtx struct {
	parent Ctx

	// compiler side
	continueBlock *ir.Block
	initBlock     *ir.Block
	forBlock      *ir.Block
	condBlock     *ir.Block
	stepBlock     *ir.Block
	breakBlock    *ir.Block
}

func (f *ForCtx) parentCtx() Ctx {
	return f.parent
}
