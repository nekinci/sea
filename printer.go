package main

import (
	"fmt"
	"io"
)

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
	Assert(w.pkg != nil, "AST is nil")
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
		_, _ = fmt.Fprintf(w.Writer, "Function[Ident: %s, Returns: %s]\n", s.Name.Name, s.Type.(*IdentExpr).Name)
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
		_, _ = fmt.Fprintf(w.Writer, "Call[%s]\n", e.Left.(*IdentExpr).Name)
		for _, param := range e.Args {
			w.indentExpr(w.writeExpr, param)
		}
	case *ParamExpr:
		_, _ = fmt.Fprintf(w.Writer, "Param[Ident: %s, Type: %s]\n", e.Name.Name, e.Type.(*IdentExpr).Name)
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
