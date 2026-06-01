package migrator

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// phase11bAutoFormat runs `pnpm exec oxfmt` to format the entire project
// using Skip's canonical formatter. This must run after phase 11 (which
// invokes pnpm install) so oxfmt is in node_modules.
//
// Reported as ok (with file count) or warn (with error tail). Never aborts —
// formatting failure is non-fatal; the user still gets the ZIP.
func phase11bAutoFormat(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "11b-auto-format"}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pnpm", "exec", "oxfmt")
	cmd.Dir = opts.OutputDir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	output := strings.TrimSpace(out.String())

	if err != nil {
		log.Status = "warn"
		log.Message = "oxfmt did not complete cleanly (skip-readiness may flag formatting)"
		log.Details = []string{truncate(output, 400)}
		log.Duration = time.Since(start).String()
		return log, nil
	}

	// Try to extract "N files" line from oxfmt's summary; otherwise just say done.
	summary := "applied oxfmt formatting"
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "files") && (strings.Contains(line, "formatted") || strings.Contains(line, "modified")) {
			summary = line
			break
		}
	}
	log.Status = "ok"
	log.Message = summary
	log.Duration = time.Since(start).String()
	return log, nil
}
