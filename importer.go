package main

import "github.com/llir/llvm/ir"

type Module struct {
	Module    *ir.Module
	TypeDefs  []*TypeDef
	FuncDef   []*FuncDef
	Globals   []*VarDef
	Imports   []*Module
	ImportMap map[string]*Module
	Scope     *Scope
	Name      string
}

func Import(path string, importMap map[string]*Module) *Module {
	pack := Parse(path)
	checker := Check(pack, importMap)
	module := Compile(pack)
	return &Module{
		Module:   module,
		TypeDefs: checker.TypeDefs,
		FuncDef:  checker.FuncDefs,
		Globals:  checker.GlobalVarDefs,
		Imports:  checker.Imports,
		Scope:    checker.initScope,
		Name:     pack.Name,
	}
}
