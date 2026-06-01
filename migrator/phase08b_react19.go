package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// React 19 changed `useRef<T>()` to require an initial argument.
// We rewrite `useRef<X>()` (no args) to `useRef<X | undefined>(undefined)`.
//
// Only applied in src/hooks/** files to limit scope; broader files (libs, pages)
// rarely use this pattern.
// Matches useRef<T>() where T may contain nested generics (e.g. ReturnType<typeof setTimeout>).
// Non-greedy capture walks character-by-character until the *first* `>()` that balances.
var useRefEmptyArgRe = regexp.MustCompile(`useRef<(.+?)>\(\)`)

// applyReact19Fixes is invoked at the end of phase 08 (alongside Zod refactor)
// to fix React 19 breaking changes in source files we copied.
func applyReact19Fixes(opts Options) (int, error) {
	root := filepath.Join(opts.OutputDir, "src", "hooks")
	if !fileExists(root) {
		return 0, nil
	}
	fixed := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".ts" && ext != ".tsx" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		original := string(data)
		if !strings.Contains(original, "useRef<") {
			return nil
		}
		updated := useRefEmptyArgRe.ReplaceAllStringFunc(original, func(match string) string {
			m := useRefEmptyArgRe.FindStringSubmatch(match)
			if len(m) < 2 {
				return match
			}
			t := m[1]
			// Avoid double-wrapping if already includes `| undefined`
			if strings.Contains(t, "undefined") {
				return match
			}
			return fmt.Sprintf("useRef<%s | undefined>(undefined)", t)
		})
		if updated != original {
			if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
				return err
			}
			fixed++
		}
		return nil
	})
	return fixed, err
}

// updatePhase08 wraps the original phase 8 to also include React 19 fixes,
// reported in the same PhaseLog for clarity.
func phase08CombinedRefactor(opts Options) (PhaseLog, error) {
	start := time.Now()
	log, err := phase08ZodRefactor(opts)
	if err != nil {
		return log, err
	}
	react19Fixed, err := applyReact19Fixes(opts)
	if err != nil {
		return log, fmt.Errorf("react19 fixes: %w", err)
	}
	if react19Fixed > 0 {
		log.Message = fmt.Sprintf("%s; react19 useRef fixed in %d hook(s)", log.Message, react19Fixed)
	}
	log.Duration = time.Since(start).String()
	return log, nil
}
