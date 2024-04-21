package main

import (
	"fmt"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var Target string
var BasePath string

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

func Parse(path string) *Package {
	dir, err2 := os.ReadDir(path)
	if err2 != nil {
		FatalPrintf("failed to read dir:  %v", err2)
	}

	pack := &Package{
		Name:      "",
		Files:     make([]*File, 0),
		FileMap:   make(map[Stmt]string),
		ImportMap: make(map[*UseStmt]string),
		Path:      path,
	}

	errors := make([]Error, 0)
	for _, item := range dir {
		if !item.IsDir() {
			if filepath.Ext(item.Name()) == ".sea" {
				join := filepath.Join(path, item.Name())
				fileContent, err2 := os.ReadFile(join)

				if err2 != nil {
					FatalPrintf("failed to read file: %v", err2)
				}

				parser := NewParser(join, string(fileContent))
				file, errors2 := parser.parse()
				pack.Name = file.PackageStmt.Name.Name
				errors = append(errors, errors2...)
				pack.Files = append(pack.Files, file)
			}
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Print(err)
		}
		ExitPrintf("")
	}

	return pack
}

func Compile(pckg *Package) *ir.Module {
	compiler := Compiler{}
	compiler.module = ir.NewModule()
	compiler.pkg = pckg
	compiler.init()
	compiler.compile()
	return compiler.module
}

func CompilePackage(path string) *Module {
	pack := Parse(path)
	checker := Check(pack, make(map[string]*Module))
	module := Compile(pack)
	return &Module{
		Module:    module,
		TypeDefs:  checker.TypeDefs,
		FuncDef:   checker.FuncDefs,
		Globals:   checker.GlobalVarDefs,
		Imports:   checker.Imports,
		Name:      pack.Name,
		ImportMap: checker.ImportMap,
	}
}

func CompileWrite(p *string, module *Module) string {
	outputPath := path.Join(os.TempDir(), module.Name+".ll")
	if p != nil {
		outputPath = *p
	}

	newFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}

	_, err = module.Module.WriteTo(newFile)
	if err != nil {
		panic(err)
	}

	return outputPath
}

func Check(pckg *Package, importMap map[string]*Module) *Checker {
	var checker = &Checker{
		Package:        pckg,
		Errors:         make([]Error, 0),
		ImportMap:      importMap,
		ImportAliasMap: make(map[string][]string),
		PathAliasMap:   make(map[string]string),
	}

	errors, err2 := checker.Check()
	if err2 {
		for _, err := range errors {
			fmt.Print(err)
		}
		ExitPrintf("")
	}

	return checker
}

func parseCommandLineArgs() {
	Target = "arm64-apple-darwin23.1.0" // TODO

	argsLen := len(os.Args)

	if argsLen < 2 || (argsLen == 2 && (os.Args[1] != "help" && os.Args[1] != "version")) {
		if argsLen < 3 {
			//binary arg, command, <file_name>
			fmt.Printf("Invalid command: %s\nExample usage: sealang run <main_package_path>", strings.Join(os.Args, " "))
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
		sealang build <main_package_path>
		sealang build <main_package_path> -o <output_path>
	run: run builds the input file and runs executable binary
	USAGE:
		sealang run <main_package_path>
	check: check checks syntax and semantic errors for given file
	USAGE:
		sealang check <main_package_path>
`)
	} else if command == "run" {
		BasePath = os.Args[2]
		_, file := path.Split(BasePath)
		CompileIt("./runtime/runtime.c", path.Join(BasePath, file), true)
	} else if command == "check" {
		pckg := Parse(os.Args[2])
		Check(pckg, make(map[string]*Module))
	} else if command == "build" {
		BasePath = os.Args[2]
		_, file := path.Split(BasePath)
		outputPath := path.Join(BasePath, file)
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
		CompileIt("./runtime/runtime.c", outputPath, false)

	}

}

func devMode() {
	Target = "arm64-apple-darwin23.1.0" // TODO
	_ = os.Setenv("DEBUG", "")
	BasePath = "./example"
	CompileIt("./runtime/runtime.c", BasePath, true)
}

func CompileIt(runtimePath string, outputPath string, runBinary bool) {
	runtimeOut := compileRuntime(runtimePath)
	pathList := make([]string, 0)
	pathList = append(pathList, runtimeOut)
	compilePackage := CompilePackage(BasePath)
	join := path.Join(BasePath, ".out")
	err := os.Mkdir(join, os.ModePerm)
	if !os.IsExist(err) {
		AssertErr(err)
	}
	o := path.Join(join, compilePackage.Name+".ll")
	wrapInitFuncs(compilePackage)
	output := CompileWrite(&o, compilePackage)
	pathList = append(pathList, output)
	for _, v := range compilePackage.ImportMap {
		o := path.Join(join, v.Name+".ll")
		write := CompileWrite(&o, v)
		pathList = append(pathList, write)
	}
	compileAll(pathList, outputPath, true, runBinary)
}

func wrapInitFuncs(compilePackage *Module) {
	initWrapperFn := compilePackage.Module.NewFunc("____INIT____", types.Void)
	block := initWrapperFn.NewBlock("entry")
	mainInit := getInitFunc("main", compilePackage.Module.Funcs)
	if mainInit != nil {
		block.NewCall(mainInit)
	}

	for _, module := range compilePackage.ImportMap {
		depInit := getInitFunc(module.Name, module.Module.Funcs)
		if depInit != nil {
			extern := compilePackage.Module.NewFunc(depInit.Name(), depInit.Sig.RetType)
			extern.Linkage = enum.LinkageExternal
			block.NewCall(depInit)
		}
	}

	block.NewRet(nil)
}

func getInitFunc(pack string, funcs []*ir.Func) *ir.Func {
	for _, fn := range funcs {
		if fn.Name() == fmt.Sprintf("__%s____init__", pack) {
			return fn
		}
	}

	return nil
}

func compileRuntime(runtimePath string) string {
	runtimeOutputPath := runtimePath[:strings.LastIndex(runtimePath, ".")] + ".ll"
	runtimeCompile := exec.Command("clang", runtimePath, "-S", "-emit-llvm", "-o", runtimeOutputPath)
	if err := runtimeCompile.Run(); err != nil {
		log.Fatalf("failed to run command for clang: %v", err)
	}
	return runtimeOutputPath
}

func main() {
	//devMode()
	parseCommandLineArgs()
}

func compileAll(pathList []string, outputPath string, outputForward bool, runBinary bool) {

	argList := make([]string, 0)
	argList = append(argList, pathList...)
	command := exec.Command("clang", argList...)
	command.Args = append(command.Args, "-o")
	command.Args = append(command.Args, outputPath)
	if outputForward {
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
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

// TODO change it
func compile(path string, abc string, runBinary bool, outputForwarding bool, outputPath string) {

	if outputPath == "" {
		outputPath = path[:strings.LastIndex(path, ".")] + ""
	}

	// clang ${path} -o ${path}.o

	pathList := make([]string, 0)
	pathList = append(pathList, path)
	compileAll(pathList, outputPath, runBinary, outputForwarding)

}
