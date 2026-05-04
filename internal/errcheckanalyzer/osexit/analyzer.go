// Package osexit provides a go/analysis Analyzer that reports calls to
// os.Exit inside func main in a package whose name is "main". Only
// non-test .go sources are considered (*_test.go files are skipped) so
// test mains may still terminate the process with os.Exit.
package osexit

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	Main       = "main" // package name and func main identifier
	PkgOS      = "os"   // import name for os.Exit selector
	FuncExit   = "Exit"
	ExtGo      = ".go"
	SuffixTest = "_test" + ExtGo // skip *_test.go
)

// Analyzer flags os.Exit(...) in func main of package main (non-test files).
// Prefer returning from main or signaling shutdown another way so defer runs.
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "forbid os.Exit inside func main in package main (excluding *_test.go)",
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
	return name == Main
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
			inspectFunc(pass, fn)
			return false
		}
		return true
	})
}
func inspectFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if id.Name == PkgOS && sel.Sel.Name == FuncExit {
			pass.Reportf(call.Pos(), "returning os.Exit")
			return false
		}
		return true
	})
}
