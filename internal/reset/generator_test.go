package reset

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteTemplate_basic(t *testing.T) {
	data := &PackageData{
		Package: "samplepkg",
		Structs: []StructInfo{
			{
				Name: "Holder",
				Fields: []StructFields{
					{Name: "Count", Type: "other", TypeName: "int"},
					{Name: "Name", Type: "other", TypeName: "string"},
					{Name: "Items", Type: "slice", TypeName: "[]byte"},
				},
			},
		},
	}
	out := executeTemplate(data)
	if !bytes.Contains(out, []byte("package samplepkg")) || !bytes.Contains(out, []byte("func (r *Holder) Reset()")) {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestGenerateFile_writesFormattedGo(t *testing.T) {
	dir := t.TempDir()
	data := &PackageData{
		Package: "gentest",
		Structs: []StructInfo{
			{
				Name: "Counter",
				Fields: []StructFields{
					{Name: "N", Type: "other", TypeName: "int"},
				},
			},
		},
	}
	GenerateFile(dir, data)
	path := filepath.Join(dir, ExtGoGen)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte("package gentest")) {
		t.Fatalf("file: %s", b)
	}
}
