package main

import (
	"fmt"
	"github.com/llir/llvm/ir/types"
)

// Custom attributes to support differences between different architectures while generating code

type PassByValue struct {
	Typ types.Type
}

func (p PassByValue) String() string {
	// As far as I know in arm64 darwin there is no byval attributes on Clang and C side.
	// So we need to generate our code appropriate to C generated convention to have advantage of C runtime library.
	if Target != "arm64-apple-darwin23.1.0" {
		return fmt.Sprintf("byval(%s)", p.Typ.String())
	}

	return ""
}

func (p PassByValue) IsParamAttribute() {}
