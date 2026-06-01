package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// phase05LibHooksTypes copies src/{lib, hooks, types, utils, assets} from source to output,
// preserving the Skip baseline's utils.ts (its cn() docstring is richer) and
// the already-present hooks (use-mobile, use-toast, useAuth).
func phase05LibHooksTypes(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "05-lib-hooks-types"}

	type job struct {
		srcSub string
		dstSub string
		skip   func(rel string) bool
	}
	jobs := []job{
		{"src/lib", "src/lib", func(rel string) bool {
			return rel == "utils.ts" // keep baseline's cn()
		}},
		{"src/hooks", "src/hooks", func(rel string) bool {
			base := filepath.Base(rel)
			return base == "use-mobile.tsx" || base == "use-toast.ts" || base == "useAuth.ts"
		}},
		{"src/types", "src/types", nil},
		{"src/utils", "src/utils", nil},
		{"src/assets", "src/assets", nil},
	}

	copied := 0
	for _, j := range jobs {
		src := filepath.Join(opts.SourceDir, j.srcSub)
		if !fileExists(src) {
			continue
		}
		dst := filepath.Join(opts.OutputDir, j.dstSub)
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return log, err
		}
		err := copyDir(src, dst, func(rel string) bool {
			if j.skip != nil && j.skip(rel) {
				return false
			}
			return true
		})
		if err != nil {
			return log, fmt.Errorf("copying %s: %w", j.srcSub, err)
		}
		copied++
	}

	log.Status = "ok"
	log.Message = fmt.Sprintf("copied %d source subdirs (lib/hooks/types/utils/assets)", copied)
	log.Duration = time.Since(start).String()
	return log, nil
}

// shouldSkipExisting is unused for now but kept for potential future use:
// returns true if the file already exists at dst (to preserve baseline content).
func shouldSkipExisting(dst string) bool {
	if !fileExists(dst) {
		return false
	}
	return strings.HasSuffix(dst, ".ts") || strings.HasSuffix(dst, ".tsx")
}
