package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOSFinder_PHPFiles_NonRecursive(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "a.php"), "<?php")
	mustWrite(t, filepath.Join(dir, "b.PHP"), "<?php")
	mustWrite(t, filepath.Join(dir, "c.txt"), "noop")
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	mustWrite(t, filepath.Join(dir, "sub", "d.php"), "<?php")

	f := NewOSFinder()
	files, err := f.PHPFiles(context.Background(), dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %#v", len(files), files)
	}
	if filepath.Base(files[0]) != "a.php" {
		t.Fatalf("unexpected first: %s", files[0])
	}
	if filepath.Base(files[1]) != "b.PHP" {
		t.Fatalf("unexpected second: %s", files[1])
	}
}

func TestOSFinder_PHPFiles_Recursive(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	mustWrite(t, filepath.Join(dir, "sub", "d.php"), "<?php")

	f := NewOSFinder()
	files, err := f.PHPFiles(context.Background(), dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %#v", len(files), files)
	}
	if filepath.Base(files[0]) != "d.php" {
		t.Fatalf("unexpected file: %s", files[0])
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
