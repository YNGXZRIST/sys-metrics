package reset

import (
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestScanPackages_findsGenerateResetStruct(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "sample.go")
	content := `package gentest

// generate:reset
type Sample struct {
	N   int
	Tag string
}
`
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	fset := token.NewFileSet()
	pkgs, err := ScanPackages(fset)
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("packages = %d", len(pkgs))
	}
	for _, data := range pkgs {
		if len(data.Structs) != 1 || data.Structs[0].Name != "Sample" {
			t.Fatalf("structs = %+v", data.Structs)
		}
	}
}
