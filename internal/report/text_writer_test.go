package report

import (
	"bytes"
	"testing"

	"php-func-parser-go/internal/model"
)

func TestTextWriter_Write(t *testing.T) {
	w := NewTextWriter()
	in := []model.FileReport{
		{Path: "file1.php", Functions: []model.FunctionDecl{{Name: "a", Params: "$x"}}},
		{Path: "file2.php", Functions: []model.FunctionDecl{{Name: "b", Params: ""}, {Name: "c", Params: "$y, $z"}}},
	}

	var buf bytes.Buffer
	if err := w.Write(&buf, in); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	want := "file1.php\nfunction a($x)\n\nfile2.php\nfunction b()\nfunction c($y, $z)\n"
	if got != want {
		t.Fatalf("unexpected output:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}
