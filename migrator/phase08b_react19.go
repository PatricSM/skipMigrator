package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Matches useRef<T>() where T may contain nested generics (e.g. ReturnType<typeof setTimeout>).
var useRefEmptyArgRe = regexp.MustCompile(`useRef<(.+?)>\(\)`)

// applyReact19Fixes rewrites useRef<X>() calls (no args) to useRef<X | undefined>(undefined).
// React 19 requires an explicit initial argument.
//
// Returns the list of files modified (relative paths) for inclusion in the report.
func applyReact19Fixes(opts Options) ([]string, error) {
	root := filepath.Join(opts.OutputDir, "src", "hooks")
	if !fileExists(root) {
		return nil, nil
	}
	var fixed []string
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
			if strings.Contains(t, "undefined") {
				return match
			}
			return fmt.Sprintf("useRef<%s | undefined>(undefined)", t)
		})
		if updated != original {
			if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
				return err
			}
			rel, _ := filepath.Rel(opts.OutputDir, path)
			fixed = append(fixed, rel)
		}
		return nil
	})
	sort.Strings(fixed)
	return fixed, err
}

// phase08CombinedRefactor wraps Zod refactor + React 19 fixes into a single phase log.
func phase08CombinedRefactor(opts Options) (PhaseLog, error) {
	start := time.Now()
	log, err := phase08ZodRefactor(opts)
	if err != nil {
		return log, err
	}
	react19Files, err := applyReact19Fixes(opts)
	if err != nil {
		return log, fmt.Errorf("react19 fixes: %w", err)
	}
	if len(react19Files) > 0 {
		log.Message = fmt.Sprintf("%s; react19 useRef corrigido em %d hook(s)", log.Message, len(react19Files))
		for _, f := range react19Files {
			log.Details = append(log.Details, fmt.Sprintf("`%s`: useRef<T>() → useRef<T | undefined>(undefined)", f))
		}
	}
	log.Duration = time.Since(start).String()
	return log, nil
}
