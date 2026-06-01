package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Zod 3 patterns that break in v4 and their replacements.
var zodReplacements = []struct {
	pattern *regexp.Regexp
	replace string
}{
	// required_error → error (parameter rename in v4)
	{regexp.MustCompile(`required_error\s*:`), `error:`},
	// invalid_type_error → error (same)
	{regexp.MustCompile(`invalid_type_error\s*:`), `error:`},
	// e.errors[0] → e.issues[0] (ZodError property rename)
	{regexp.MustCompile(`\.errors\[`), `.issues[`},
	// err?.errors?.[ → err?.issues?.[
	{regexp.MustCompile(`\?\.errors\?\.\[`), `?.issues?.[`},
	// errorMap → error callback (best-effort, may need manual review)
	{regexp.MustCompile(`errorMap\s*:`), `error:`},
}

// phase08ZodRefactor walks OutputDir's TypeScript files and applies Zod 3→4 substitutions.
func phase08ZodRefactor(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "08-zod-refactor"}

	root := filepath.Join(opts.OutputDir, "src")
	changedFiles := 0
	totalReplacements := 0

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

		// Quick skip: only process files that import zod
		if !strings.Contains(original, "from 'zod'") &&
			!strings.Contains(original, `from "zod"`) &&
			!strings.Contains(original, "ZodError") &&
			!strings.Contains(original, ".errors[") &&
			!strings.Contains(original, "?.errors?.[") {
			return nil
		}

		updated := original
		fileReplacements := 0
		for _, r := range zodReplacements {
			matches := r.pattern.FindAllStringIndex(updated, -1)
			if len(matches) > 0 {
				updated = r.pattern.ReplaceAllString(updated, r.replace)
				fileReplacements += len(matches)
			}
		}

		if updated != original {
			if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
				return err
			}
			changedFiles++
			totalReplacements += fileReplacements
		}
		return nil
	})
	if err != nil {
		return log, fmt.Errorf("walk: %w", err)
	}

	log.Status = "ok"
	log.Message = fmt.Sprintf("zod3→4: %d replacements across %d files", totalReplacements, changedFiles)
	log.Duration = time.Since(start).String()
	return log, nil
}
