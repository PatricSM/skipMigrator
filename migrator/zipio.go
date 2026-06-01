package migrator

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UnzipToDir extracts zipPath into dst. The top-level dir of the archive
// (if every entry shares the same first path segment) is stripped.
func UnzipToDir(zipPath, dst string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("opening zip: %w", err)
	}
	defer r.Close()

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return "", err
	}

	stripPrefix := commonPrefix(r.File)

	for _, f := range r.File {
		name := f.Name
		if stripPrefix != "" && strings.HasPrefix(name, stripPrefix) {
			name = strings.TrimPrefix(name, stripPrefix)
		}
		if name == "" {
			continue
		}
		target := filepath.Join(dst, name)
		// Guard against zip-slip
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dst)+string(os.PathSeparator)) {
			return "", fmt.Errorf("zip-slip detected: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return "", err
		}
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}
		in, err := f.Open()
		if err != nil {
			out.Close()
			return "", err
		}
		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			out.Close()
			return "", err
		}
		in.Close()
		out.Close()
	}
	return dst, nil
}

// ZipDir packages srcDir into zipPath. Entries inside the archive are relative to srcDir
// (no extra prefix), and node_modules / .git / lockfile cache dirs are skipped.
func ZipDir(srcDir, zipPath string) error {
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		// Skip noisy / huge dirs
		if info.IsDir() {
			base := info.Name()
			if base == "node_modules" || base == ".git" || base == "dist" || base == "dev-dist" || base == ".pnpm-store" || base == ".pnpm" {
				return filepath.SkipDir
			}
			return nil
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		hdr.Method = zip.Deflate
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(w, in)
		return err
	})
}

// commonPrefix returns the shared top-level path segment across all entries (with trailing slash),
// or "" if there is no shared prefix.
func commonPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}
	first := strings.SplitN(files[0].Name, "/", 2)
	if len(first) < 2 {
		return ""
	}
	prefix := first[0] + "/"
	for _, f := range files {
		if !strings.HasPrefix(f.Name, prefix) {
			return ""
		}
	}
	return prefix
}
