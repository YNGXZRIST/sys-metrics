package main

import (
	"go/token"
	"sys-metrics/internal/reset"
)

func main() {
	fset := token.NewFileSet()
	pkgMap, err := reset.ScanPackages(fset)
	if err != nil {
		panic(err)
	}
	for dir, data := range pkgMap {
		reset.GenerateFile(dir, data)
	}
}
