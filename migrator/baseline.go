package migrator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:skipbase
var skipbaseFS embed.FS

// WriteBaseline expands the embedded Skip starter template into outDir.
// The embedded files live under skipbase/; they are written to outDir without that prefix.
func WriteBaseline(outDir string) error {
	return fs.WalkDir(skipbaseFS, "skipbase", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "skipbase" {
			return nil
		}
		rel := path[len("skipbase/"):]
		dst := filepath.Join(outDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}
		data, err := skipbaseFS.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("writing baseline file %s: %w", rel, err)
		}
		return nil
	})
}
