package reset

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func mustParseFile(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return fset, node
}

func mustGetStructType(t *testing.T, node *ast.File, name string) *ast.StructType {
	t.Helper()
	for _, decl := range node.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != name {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				t.Fatalf("%s is not a struct", name)
			}
			return st
		}
	}
	t.Fatalf("struct %s not found", name)
	return nil
}

func TestIsSkippedFile(t *testing.T) {
	tests := []struct {
		name    string
		skipped bool
	}{
		{"main.go", false},
		{".hidden.go", true},
		{"foo_test.go", true},
		{"reset.gen.go", false},
		{"walker.go", false},
		{"foo.txt", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSkippedFile(tt.name)
			if got != tt.skipped {
				t.Errorf("isSkippedFile(%q) = %v, want %v", tt.name, got, tt.skipped)
			}
		})
	}
}

func TestFindResetStructs_NoMarker(t *testing.T) {
	fset, node := mustParseFile(t, `package foo

type Foo struct {
	X int
}
`)
	structs, err := findResetStructs(fset, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(structs) != 0 {
		t.Errorf("expected 0 structs, got %d", len(structs))
	}
}

func TestFindResetStructs_MarkerDirectlyAbove(t *testing.T) {
	fset, node := mustParseFile(t, `package foo

// generate:reset
type Foo struct {
	X int
}
`)
	structs, err := findResetStructs(fset, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(structs) != 1 || structs[0].Name != "Foo" {
		t.Errorf("expected [Foo], got %v", structs)
	}
}

func TestFindResetStructs_MarkerWithDocComment(t *testing.T) {
	fset, node := mustParseFile(t, `package foo

// generate:reset

// Foo is a struct.
type Foo struct {
	X int
}
`)
	structs, err := findResetStructs(fset, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(structs) != 1 || structs[0].Name != "Foo" {
		t.Errorf("expected [Foo], got %v", structs)
	}
}

func TestFindResetStructs_MultipleMarkers(t *testing.T) {
	fset, node := mustParseFile(t, `package foo

// generate:reset
type Foo struct { X int }

// generate:reset
type Bar struct { S string }

type Baz struct { N float64 }
`)
	structs, err := findResetStructs(fset, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(structs) != 2 {
		t.Fatalf("expected 2 structs, got %d", len(structs))
	}
	if structs[0].Name != "Foo" {
		t.Errorf("expected Foo, got %s", structs[0].Name)
	}
	if structs[1].Name != "Bar" {
		t.Errorf("expected Bar, got %s", structs[1].Name)
	}
}

func TestClassifyField(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		structName   string
		fieldName    string
		wantKind     string
		wantTypeName string
	}{
		{
			name:       "int",
			src:        `package foo; type T struct { X int }`,
			structName: "T", fieldName: "X",
			wantKind: "other", wantTypeName: "int",
		},
		{
			name:       "string",
			src:        `package foo; type T struct { S string }`,
			structName: "T", fieldName: "S",
			wantKind: "other", wantTypeName: "string",
		},
		{
			name:       "bool",
			src:        `package foo; type T struct { B bool }`,
			structName: "T", fieldName: "B",
			wantKind: "other", wantTypeName: "bool",
		},
		{
			name:       "selector sync.Mutex",
			src:        `package foo; import "sync"; type T struct { M sync.Mutex }`,
			structName: "T", fieldName: "M",
			wantKind: "other", wantTypeName: "sync.Mutex",
		},
		{
			name:       "pointer",
			src:        `package foo; type T struct { P *int }`,
			structName: "T", fieldName: "P",
			wantKind: "pointer", wantTypeName: "*int",
		},
		{
			name:       "slice",
			src:        `package foo; type T struct { Sl []string }`,
			structName: "T", fieldName: "Sl",
			wantKind: "slice", wantTypeName: "[]string",
		},
		{
			name:       "map",
			src:        `package foo; type T struct { M map[string]int }`,
			structName: "T", fieldName: "M",
			wantKind: "map", wantTypeName: "map[string]int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset, node := mustParseFile(t, tt.src)
			st := mustGetStructType(t, node, tt.structName)

			var fieldExpr ast.Expr
			for _, f := range st.Fields.List {
				for _, n := range f.Names {
					if n.Name == tt.fieldName {
						fieldExpr = f.Type
					}
				}
			}
			if fieldExpr == nil {
				t.Fatalf("field %s not found", tt.fieldName)
			}

			kind, typeName, err := classifyField(fset, fieldExpr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if kind != tt.wantKind {
				t.Errorf("kind: got %q, want %q", kind, tt.wantKind)
			}
			if typeName != tt.wantTypeName {
				t.Errorf("typeName: got %q, want %q", typeName, tt.wantTypeName)
			}
		})
	}
}

func TestCollectFields_SkipsEmbedded(t *testing.T) {
	fset, node := mustParseFile(t, `package foo

import "sync"

type Foo struct {
	sync.Mutex
	X int
	S string
}
`)
	st := mustGetStructType(t, node, "Foo")
	fields, err := collectFields(fset, st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields (embedded skipped), got %d", len(fields))
	}
	if fields[0].Name != "X" {
		t.Errorf("expected X, got %s", fields[0].Name)
	}
	if fields[1].Name != "S" {
		t.Errorf("expected S, got %s", fields[1].Name)
	}
}
