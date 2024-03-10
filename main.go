package main

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"log"
	"os"
	"os/exec"
	"strings"
)

var Target string

func Assert(cond bool, msg string) {
	if cond {
		return
	}
	panic(msg)
}

func AssertErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main2() {
	module := ir.NewModule()
	newStruct := types.NewStruct(types.I32, types.I32)
	typ := module.NewTypeDef("deneme", newStruct)
	_ = typ
	str := constant.NewStruct(newStruct)
	str.Fields = append(str.Fields, constant.NewInt(types.I32, 3))
	str.Fields = append(str.Fields, constant.NewInt(types.I32, 5))
	f := module.NewFunc("main", types.I32)
	block := f.NewBlock("entry")
	alloca := block.NewAlloca(typ)
	block.NewStore(str, alloca)
	block.NewRet(constant.NewInt(types.I32, 0))
	module.WriteTo(os.Stdout)
}

func main() {
	_ = os.Setenv("DEBUG", "")

	Target = "arm64-apple-darwin23.1.0" // TODO
	compiler := &Compiler{}

	// read compileExpr from file
	file, err := os.ReadFile("./input.sea")
	parser := NewParser("./input.sea")
	//parser.printTokens()
	pckg, errors := parser.parse()
	if errors != nil && len(errors) > 0 {
		for _, err := range errors {
			fmt.Print(err)
		}
		return
	}

	var checker *Checker = &Checker{
		Package: pckg,
		Errors:  make([]Error, 0),
	}

	errors, err2 := checker.Check()
	if err2 {
		for _, err := range errors {
			fmt.Print(err)
		}
		return
	}

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
