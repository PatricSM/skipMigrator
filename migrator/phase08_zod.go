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

// Zod 3 patterns that break in v4 and their replacements.
var zodReplacements = []struct {
	name    string
	pattern *regexp.Regexp
	replace string
}{
	{"required_error→error", regexp.MustCompile(`required_error\s*:`), `error:`},
	{"invalid_type_error→error", regexp.MustCompile(`invalid_type_error\s*:`), `error:`},
	{".errors[→.issues[", regexp.MustCompile(`\.errors\[`), `.issues[`},
	{"?.errors?.[→?.issues?.[", regexp.MustCompile(`\?\.errors\?\.\[`), `?.issues?.[`},
	{"errorMap→error", regexp.MustCompile(`errorMap\s*:`), `error:`},
}

// phase08ZodRefactor walks OutputDir's TypeScript files and applies Zod 3→4 substitutions.
// Returns a PhaseLog with per-file details populated for the report.
func phase08ZodRefactor(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "08-zod-refactor"}

	root := filepath.Join(opts.OutputDir, "src")
	changedFiles := map[string]map[string]int{} // file → {ruleName: count}
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
		if !strings.Contains(original, "from 'zod'") &&
			!strings.Contains(original, `from "zod"`) &&
			!strings.Contains(original, "ZodError") &&
			!strings.Contains(original, ".errors[") &&
			!strings.Contains(original, "?.errors?.[") {
			return nil
		}

		updated := original
		fileTotals := map[string]int{}
		for _, r := range zodReplacements {
			matches := r.pattern.FindAllStringIndex(updated, -1)
			if len(matches) > 0 {
				updated = r.pattern.ReplaceAllString(updated, r.replace)
				fileTotals[r.name] += len(matches)
				totalReplacements += len(matches)
			}
		}

		if updated != original {
			if err := os.WriteFile(path, []byte(updated), info.Mode()); err != nil {
				return err
			}
			rel, _ := filepath.Rel(opts.OutputDir, path)
			changedFiles[rel] = fileTotals
		}
		return nil
	})
	if err != nil {
		return log, fmt.Errorf("walk: %w", err)
	}

	// Compose details (sorted by file path for stability)
	files := make([]string, 0, len(changedFiles))
	for f := range changedFiles {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		parts := []string{}
		for rule, n := range changedFiles[f] {
			parts = append(parts, fmt.Sprintf("%s (%d×)", rule, n))
		}
		sort.Strings(parts)
		log.Details = append(log.Details, fmt.Sprintf("`%s`: %s", f, strings.Join(parts, ", ")))
	}

	log.Status = "ok"
	log.Message = fmt.Sprintf("zod3→4: %d substituições em %d arquivos", totalReplacements, len(changedFiles))
	log.Duration = time.Since(start).String()
	return log, nil
}
