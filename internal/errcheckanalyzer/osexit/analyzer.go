// Package osexit provides a go/analysis Analyzer that reports calls to
// os.Exit, log.Fatal*, and panic inside func main in a package whose name is "main".
// Only non-test .go sources are considered (*_test.go files are skipped) so
// test mains may still terminate the process.
package osexit

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	Main        = "main" // package name and func main identifier
	Run         = "run"  //package run main func identifier
	PkgOS       = "os"   // import name for os.Exit selector
	FuncExit    = "Exit"
	PkgLog      = "log" // import name for log.Fatal selector
	FuncFatal   = "Fatal"
	FuncFatalf  = "Fatalf"
	FuncFatalln = "Fatalln"
	FuncPanic   = "panic" // builtin
	ExtGo       = ".go"
	SuffixTest  = "_test" + ExtGo // skip *_test.go
)

// Analyzer flags os.Exit(...), log.Fatal*(...), and panic(...) in func main of package main (non-test files).
// Prefer returning from main or signaling shutdown another way so defer runs.
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "forbid os.Exit/log.Fatal*/panic inside func main in package main (excluding *_test.go)",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if !isPerformingFile(pass, file) {
			continue
		}
		if !isMainPkg(file) {
			continue
		}
		inspectFile(pass, file)
	}
	return nil, nil
}
func isPerformingFile(pass *analysis.Pass, file *ast.File) bool {
	fileName := pass.Fset.Position(file.Pos()).Filename
	if filepath.Ext(fileName) != ExtGo || strings.HasSuffix(fileName, SuffixTest) {
		return false
	}
	return true
}
func isMainPkg(file *ast.File) bool {
	pkgName := file.Name.Name
	return isMain(pkgName)
}
func isMain(name string) bool {
	return name == Main || name == Run
}
func isMainFunc(fn *ast.FuncDecl) bool {
	return isMain(fn.Name.Name)
}
func inspectFile(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if !isMainFunc(fn) {
				return true
			}
			inspectMainFunc(pass, fn)
			return false
		}
		return true
	})
}
func inspectMainFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if id, okIdent := call.Fun.(*ast.Ident); okIdent && id.Name == FuncPanic {
			pass.Reportf(call.Pos(), "returning panic")
			return false
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		recv, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if recv.Name == PkgOS && sel.Sel.Name == FuncExit {
			pass.Reportf(call.Pos(), "returning os.Exit")
			return false
		}
		if recv.Name == PkgLog && (sel.Sel.Name == FuncFatal || sel.Sel.Name == FuncFatalf || sel.Sel.Name == FuncFatalln) {
			pass.Reportf(call.Pos(), "returning log.%s", sel.Sel.Name)
			return false
		}
		return true
	})
}
