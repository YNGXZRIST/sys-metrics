package main

import (
	"fmt"
	"go/token"
	"log"
	"sys-metrics/internal/reset"
)

func main() {
	if err := run(); err != nil {
		log.Printf("fatal: %s", err)
		return
	}
}
func run() error {
	fset := token.NewFileSet()
	pkgMap, err := reset.ScanPackages(fset)
	if err != nil {
		return fmt.Errorf("scanning packages: %w", err)
	}
	for dir, data := range pkgMap {
		reset.GenerateFile(dir, data)
	}
	return nil
}
