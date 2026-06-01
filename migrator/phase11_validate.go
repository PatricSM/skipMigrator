package migrator

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// phase11Validate runs pnpm install + tsc --noEmit + pnpm build in OutputDir.
// Returns the combined stdout/stderr in buildLog. On failure, the migration is considered
// not validated; the caller decides whether to return the ZIP anyway.
func phase11Validate(opts Options) (PhaseLog, string, error) {
	start := time.Now()
	log := PhaseLog{Phase: "11-validate"}

	var combined bytes.Buffer

	steps := []struct {
		name string
		args []string
	}{
		{"pnpm install", []string{"install", "--frozen-lockfile=false", "--reporter=ndjson"}},
		{"tsc --noEmit", []string{"exec", "tsc", "--noEmit", "-p", "tsconfig.app.json"}},
		{"pnpm build", []string{"build"}},
	}

	for _, step := range steps {
		combined.WriteString(fmt.Sprintf("\n$ pnpm %s\n", strings.Join(step.args, " ")))
		cmd := exec.Command("pnpm", step.args...)
		cmd.Dir = opts.OutputDir
		cmd.Stdout = &combined
		cmd.Stderr = &combined
		if err := cmd.Run(); err != nil {
			log.Status = "error"
			log.Message = fmt.Sprintf("%s failed: %v", step.name, err)
			log.Duration = time.Since(start).String()
			return log, combined.String(), fmt.Errorf("%s failed: %w", step.name, err)
		}
	}

	log.Status = "ok"
	log.Message = "validation passed (install + tsc + build)"
	log.Duration = time.Since(start).String()
	return log, combined.String(), nil
}
