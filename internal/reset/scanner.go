// Package reset scans Go source files for structs marked with // generate:reset and generates Reset() methods for them.
package reset

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	Root       = "."
	Hidden     = "."
	ExtGo      = ".go"
	SuffixTest = "_test" + ExtGo
	Marker     = "// generate:reset"
)
const (
	Other   = "other"
	Pointer = "pointer"
	Slice   = "slice"
	Map     = "map"
	Any     = "any"
)
const (
	FnReset = "Reset"
)

type StructInfo struct {
	Name   string
	Fields []StructFields
}

type StructFields struct {
	Name     string
	Type     string
	TypeName string
}

type PackageData struct {
	hasResetFor map[string]bool
	Package     string
	Structs     []StructInfo
}

func ScanPackages(fset *token.FileSet) (map[string]*PackageData, error) {
	pkgMap := map[string]*PackageData{}

	err := filepath.WalkDir(Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if isSkippedFile(d.Name()) {
			return nil
		}
		return processFile(fset, path, pkgMap)
	})
	if err != nil {
		return nil, fmt.Errorf("walk dir: %w", err)
	}
	return pkgMap, nil
}

func isSkippedFile(name string) bool {
	return strings.HasPrefix(name, Hidden) ||
		!strings.HasSuffix(name, ExtGo) ||
		strings.HasSuffix(name, SuffixTest)
}

func processFile(fset *token.FileSet, path string, pkgMap map[string]*PackageData) error {
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse error %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	if _, ok := pkgMap[dir]; !ok {
		pkgMap[dir] = &PackageData{Package: node.Name.Name, hasResetFor: map[string]bool{}}
	}

	for _, name := range existingResetMethods(node) {
		pkgMap[dir].hasResetFor[name] = true
	}

	structs, err := findResetStructs(fset, node)
	if err != nil {
		return fmt.Errorf("find in struct errors %s: %w", path, err)
	}

	for _, s := range structs {
		if !pkgMap[dir].hasResetFor[s.Name] {
			pkgMap[dir].Structs = append(pkgMap[dir].Structs, s)
		}
	}
	return nil
}

func existingResetMethods(node *ast.File) []string {
	var names []string
	for _, decl := range node.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != FnReset || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		recv := fn.Recv.List[0].Type
		if star, ok := recv.(*ast.StarExpr); ok {
			recv = star.X
		}
		if ident, ok := recv.(*ast.Ident); ok {
			names = append(names, ident.Name)
		}
	}
	return names
}

func findResetStructs(fset *token.FileSet, node *ast.File) ([]StructInfo, error) {
	typeDecls := collectTypeDecls(node)

	var result []StructInfo
	for _, cg := range node.Comments {
		for _, c := range cg.List {
			if strings.TrimSpace(c.Text) != Marker {
				continue
			}
			info, ok, err := nextStructAfter(fset, c.Pos(), typeDecls)
			if err != nil {
				return nil, err
			}
			if ok {
				result = append(result, info)
			}
		}
	}
	return result, nil
}

func collectTypeDecls(node *ast.File) []*ast.GenDecl {
	var decls []*ast.GenDecl
	for _, decl := range node.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			decls = append(decls, gd)
		}
	}
	return decls
}

func nextStructAfter(fset *token.FileSet, pos token.Pos, typeDecls []*ast.GenDecl) (StructInfo, bool, error) {
	for _, decl := range typeDecls {
		if decl.Pos() <= pos {
			continue
		}
		for _, spec := range decl.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			str, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			fields, err := collectFields(fset, str)
			if err != nil {
				return StructInfo{}, false, fmt.Errorf("nextStructAfter %s: %w", ts.Name.Name, err)
			}
			return StructInfo{Name: ts.Name.Name, Fields: fields}, true, nil
		}
		break
	}
	return StructInfo{}, false, nil
}

func collectFields(fset *token.FileSet, str *ast.StructType) ([]StructFields, error) {
	fields := make([]StructFields, 0, len(str.Fields.List))
	for _, field := range str.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		kind, typeName, err := classifyField(fset, field.Type)
		if err != nil {
			return nil, err
		}
		for _, name := range field.Names {
			fields = append(fields, StructFields{
				Name:     name.Name,
				Type:     kind,
				TypeName: typeName,
			})
		}
	}
	return fields, nil
}

func classifyField(fset *token.FileSet, expr ast.Expr) (kind, typeName string, err error) {
	switch typ := expr.(type) {
	case *ast.Ident:
		return Other, typ.Name, nil
	case *ast.SelectorExpr:
		return Other, fmt.Sprintf("%s.%s", typ.X, typ.Sel.Name), nil
	case *ast.StarExpr:
		s, err := nodeToString(fset, typ)
		return Pointer, s, err
	case *ast.ArrayType:
		s, err := nodeToString(fset, typ)
		return Slice, s, err
	case *ast.MapType:
		s, err := nodeToString(fset, typ)
		return Map, s, err
	case *ast.IndexExpr, *ast.IndexListExpr:
		s, err := nodeToString(fset, expr)
		return Other, s, err
	case *ast.ChanType:
		s, err := nodeToString(fset, typ)
		return Other, s, err
	case *ast.InterfaceType:
		return Other, Any, nil
	default:
		s, _ := nodeToString(fset, expr)
		return Other, s, nil
	}
}

func nodeToString(fset *token.FileSet, node ast.Node) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
