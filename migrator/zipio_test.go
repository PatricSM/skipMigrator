package migrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestZipRoundTrip writes a small tree, zips it, unzips elsewhere, asserts equality.
func TestZipRoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	mustWrite(t, filepath.Join(src, "a.txt"), "hello")
	mustWrite(t, filepath.Join(src, "sub", "b.txt"), "world")
	// node_modules should be skipped during zip
	mustWrite(t, filepath.Join(src, "node_modules", "leaked.txt"), "should-not-appear")

	zipPath := filepath.Join(dir, "out.zip")
	if err := ZipDir(src, zipPath); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "unzipped")
	if _, err := UnzipToDir(zipPath, dst); err != nil {
		t.Fatal(err)
	}

	a, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(a) != "hello" {
		t.Errorf("a.txt = %q", a)
	}
	b, _ := os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if string(b) != "world" {
		t.Errorf("sub/b.txt = %q", b)
	}
	if _, err := os.Stat(filepath.Join(dst, "node_modules", "leaked.txt")); err == nil {
		t.Error("node_modules should have been skipped by ZipDir")
	}
}

// TestUnzipStripsCommonPrefix handles the wildly common "github-style" ZIP
// where every entry shares a top dir (e.g. `my-repo-main/...`).
func TestUnzipStripsCommonPrefix(t *testing.T) {
	src := t.TempDir()
	inner := filepath.Join(src, "my-repo-main")
	mustWrite(t, filepath.Join(inner, "README.md"), "hi")
	mustWrite(t, filepath.Join(inner, "src", "main.ts"), "code")

	// Write zip OUTSIDE the src tree so it doesn't get recursively included.
	zipPath := filepath.Join(t.TempDir(), "src.zip")
	if err := ZipDir(src, zipPath); err != nil {
		t.Fatal(err)
	}

	dst := t.TempDir()
	if _, err := UnzipToDir(zipPath, dst); err != nil {
		t.Fatal(err)
	}
	// The common prefix "my-repo-main/" should have been stripped on unzip.
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Errorf("expected README.md at top of unzip dir, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "src", "main.ts")); err != nil {
		t.Errorf("expected src/main.ts at top of unzip dir, got %v", err)
	}
}

// TestCommonPrefix exercises the helper directly on a few archive layouts.
func TestCommonPrefix_NoSharedRoot(t *testing.T) {
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "README.md"), "hi")          // top-level
	mustWrite(t, filepath.Join(src, "sub", "b.txt"), "world")    // nested

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := ZipDir(src, zipPath); err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	if _, err := UnzipToDir(zipPath, dst); err != nil {
		t.Fatal(err)
	}
	// No shared top dir → nothing stripped, README.md still at root, sub/ preserved.
	if _, err := os.Stat(filepath.Join(dst, "README.md")); err != nil {
		t.Errorf("README.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "b.txt")); err != nil {
		t.Errorf("sub/b.txt missing: %v", err)
	}
}

// unused import guard
var _ = strings.HasPrefix
