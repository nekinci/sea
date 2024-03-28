package main

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"log"
	"os"
	"os/exec"
	"path"
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

func ExitPrintf(format string, args ...any) {
	if len(args) > 0 {
		fmt.Printf(format, args)
	} else {
		fmt.Print(format)
	}
	os.Exit(0)
}

func FatalPrintf(format string, args ...any) {
	fmt.Printf(format, args)
	os.Exit(1)
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

func Parse(path string) *Package {
	parser := NewParser(path)
	pckg, errors := parser.parse()
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Print(err)
		}
		ExitPrintf("")
	}

	return pckg
}

func Compile(p *string, pckg *Package) string {
	outputPath := path.Join(os.TempDir(), "plus.ll")
	if p != nil {
		outputPath = *p
	}

	newFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}

	compiler := Compiler{}
	compiler.module = ir.NewModule()
	compiler.init()
	compiler.pkg = pckg
	compiler.compile()
	_, err = compiler.module.WriteTo(newFile)
	if err != nil {
		panic(err)
	}

	return outputPath
}

func Check(pckg *Package) {
	var checker = &Checker{
		Package: pckg,
		Errors:  make([]Error, 0),
	}

	errors, err2 := checker.Check()
	if err2 {
		for _, err := range errors {
			fmt.Print(err)
		}
		ExitPrintf("")
	}
}

func parseCommandLineArgs() {
	Target = "arm64-apple-darwin23.1.0" // TODO

	argsLen := len(os.Args)

	if argsLen < 2 || (argsLen == 2 && (os.Args[1] != "help" && os.Args[1] != "version")) {
		if argsLen < 3 {
			//binary arg, command, <file_name>
			fmt.Printf("Invalid command: %s\nExample usage: sealang run <filename>", strings.Join(os.Args, " "))
			os.Exit(1)
		}
	}

	command := os.Args[1]

	if command == "version" {
		ExitPrintf("sealang version 0.0.1\n")
	} else if command == "help" {
		fmt.Fprintf(os.Stdout, `SEALANG v0.0.1
sea is a toy programming language in C family pronounced as C :)
DISCLAIMER: This is the early unstable version
commands:
	help: help prints useful command tips
	build: build compiles input and runtime files and extracts executable binary to current directory or given output parameter path
	USAGE:
		sealang build <file_name>
		sealang build <file_name> -o <output_path>
	run: run builds the input file and runs executable binary
	USAGE:
		sealang run <file_name>
	check: check checks syntax and semantic errors for given file
	USAGE:
		sealang check <file_name>
`)
	} else if command == "run" {
		pckg := Parse(os.Args[2])
		Check(pckg)
		output := Compile(nil, pckg)
		compile(output, "./runtime/runtime.c", true, true, "")
	} else if command == "check" {
		pckg := Parse(os.Args[2])
		Check(pckg)
	} else if command == "build" {
		outputPath := ""
		if argsLen > 3 {
			argv3 := os.Args[3]
			if argv3 != "-o" {
				ExitPrintf("invalid parameter: %s\nExample usage: ... -o output_path ", argv3)
			}
			if argsLen != 5 {
				ExitPrintf("invalid parameters: %s\n", "")
			}
			outputPath = os.Args[4]
		}
		pckg := Parse(os.Args[2])
		Check(pckg)
		output := Compile(nil, pckg)
		compile(output, "./runtime/runtime.c", false, false, outputPath)

	}

}

func devMode() {
	Target = "arm64-apple-darwin23.1.0" // TODO
	_ = os.Setenv("DEBUG", "")
	// read compileExpr from file
	pckg := Parse("./input.sea")
	Check(pckg)
	outputLL := "./input.ll"
	output := Compile(&outputLL, pckg)
	compile(output, "./runtime/runtime.c", true, true, "")
}

func main() {
	//	devMode()
	parseCommandLineArgs()
}

// TODO change it
func compile(path, runtimePath string, runBinary bool, outputForwarding bool, outputPath string) {

	if outputPath == "" {
		outputPath = path[:strings.LastIndex(path, ".")] + ""
	}

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
