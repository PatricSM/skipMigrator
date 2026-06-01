// Package migrator orchestrates the Lovable → Skip migration.
// Run via the CLI in cmd/migrate-cli or programmatically via Run().
package migrator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Options controls migration behavior.
type Options struct {
	// SourceDir: path to extracted Lovable project (input)
	SourceDir string
	// OutputDir: path where the migrated Skip project will be written
	OutputDir string
	// SupabaseURL / SupabaseAnonKey: optional overrides; otherwise extracted from source
	SupabaseURL     string
	SupabaseAnonKey string
	// PixelPerfect: when true, overrides components/ui with source versions and adapts calendar v8→v9
	PixelPerfect bool
	// Validate: when true, runs `pnpm install + tsc --noEmit + pnpm build` in OutputDir as the final step
	Validate bool
	// LogFn: optional log callback; defaults to printing to stderr
	LogFn func(format string, args ...any)
}

// Result captures the outcome of a migration run.
type Result struct {
	OutputDir       string
	PhaseLogs       []PhaseLog
	BuildLog        string
	Warnings        []string
	SupabaseURL     string // URL actually used (extracted or overridden)
	SupabaseAnonKey string
}

// PhaseLog records the outcome of a single phase.
type PhaseLog struct {
	Phase    string
	Status   string // "ok" | "skipped" | "warn" | "error"
	Message  string
	Duration string
}

// Run executes all 11 phases of the Lovable → Skip migration.
// The output directory is populated from scratch using the embedded Skip baseline,
// then progressively overlaid with content from SourceDir.
func Run(opts Options) (*Result, error) {
	if opts.LogFn == nil {
		opts.LogFn = func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}
	}
	if opts.SourceDir == "" {
		return nil, fmt.Errorf("Options.SourceDir is required")
	}
	if opts.OutputDir == "" {
		return nil, fmt.Errorf("Options.OutputDir is required")
	}

	if err := ensureEmptyDir(opts.OutputDir); err != nil {
		return nil, fmt.Errorf("preparing output dir: %w", err)
	}

	result := &Result{OutputDir: opts.OutputDir}

	// Lay down the Skip baseline first; phases mutate it.
	if err := WriteBaseline(opts.OutputDir); err != nil {
		return nil, fmt.Errorf("writing Skip baseline: %w", err)
	}
	opts.LogFn("Skip baseline written to %s", opts.OutputDir)

	pipeline := []phaseFn{
		phase01Preflight,
		phase02Dependencies,
		phase03SupabaseClient,
		phase04Auth,
		phase05LibHooksTypes,
		phase06Components,
		phase07Pages,
		phase08CombinedRefactor,
		phase09BuildConfigs,
		phase10SupabaseVersioning,
	}

	for _, fn := range pipeline {
		log, err := fn(opts)
		result.PhaseLogs = append(result.PhaseLogs, log)
		opts.LogFn("[%s] %s — %s", log.Phase, log.Status, log.Message)
		if err != nil {
			return result, fmt.Errorf("%s failed: %w", log.Phase, err)
		}
	}

	if opts.Validate {
		log, buildOut, err := phase11Validate(opts)
		result.PhaseLogs = append(result.PhaseLogs, log)
		result.BuildLog = buildOut
		opts.LogFn("[%s] %s — %s", log.Phase, log.Status, log.Message)
		if err != nil {
			return result, fmt.Errorf("%s failed: %w", log.Phase, err)
		}
	}

	return result, nil
}

// phaseFn is the signature each migration phase implements.
type phaseFn func(opts Options) (PhaseLog, error)

// ensureEmptyDir creates dir if it doesn't exist; if it does and is non-empty, fails to avoid clobbering.
func ensureEmptyDir(dir string) error {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists and is not a directory", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("output dir %s is not empty (refusing to overwrite)", dir)
	}
	return nil
}

// copyFile copies src to dst, creating parent dirs as needed.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// Preserve mode
	if info, err := in.Stat(); err == nil {
		_ = os.Chmod(dst, info.Mode())
	}
	return nil
}

// copyDir recursively copies src directory tree into dst.
// If filter returns false for a path, it's skipped.
func copyDir(src, dst string, filter func(relPath string) bool) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if filter != nil && !filter(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath)
	})
}

// fileExists reports whether the named file or directory exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
