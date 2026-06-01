package migrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// phase13SkipReadiness verifies the output conforms to Skip platform conventions:
//
//   - oxlint passes (errors only; warnings are tolerated)
//   - oxfmt --check passes
//   - .skip.config.json parses as JSON
//   - @skip-protected markers present in index.html and src/main.tsx
//   - vite-plugin-react-uid.js present at root
//   - canonical scripts present in package.json (start, dev, build, lint, format)
//
// Any failure is recorded as a single warn in the phase log so the user sees it
// in MIGRATION_REPORT.md, but the migration is NOT aborted: the ZIP still
// downloads (the user can fix and re-run, or push as-is).
func phase13SkipReadiness(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "13-skip-readiness"}

	var failures []string

	// 1. oxlint (errors only)
	if msg, ok := runOxCheck(opts.OutputDir, "oxlint", []string{"src"}); !ok {
		failures = append(failures, "oxlint: "+msg)
	}

	// 2. oxfmt --check
	if msg, ok := runOxCheck(opts.OutputDir, "oxfmt", []string{"--check"}); !ok {
		failures = append(failures, "oxfmt: "+msg)
	}

	// 3. .skip.config.json valid JSON
	if data, err := os.ReadFile(filepath.Join(opts.OutputDir, ".skip.config.json")); err != nil {
		failures = append(failures, ".skip.config.json: missing ("+err.Error()+")")
	} else {
		var dummy any
		if err := json.Unmarshal(data, &dummy); err != nil {
			failures = append(failures, ".skip.config.json: invalid JSON: "+err.Error())
		}
	}

	// 4. @skip-protected markers
	if msg, ok := checkSkipProtectedMarkers(opts.OutputDir); !ok {
		failures = append(failures, msg)
	}

	// 5. vite-plugin-react-uid.js present
	if !fileExists(filepath.Join(opts.OutputDir, "vite-plugin-react-uid.js")) {
		failures = append(failures, "vite-plugin-react-uid.js: missing from project root")
	}

	// 6. canonical package.json scripts
	if msg, ok := checkPackageScripts(opts.OutputDir); !ok {
		failures = append(failures, "package.json: "+msg)
	}

	if len(failures) == 0 {
		log.Status = "ok"
		log.Message = "Skip conventions satisfied (oxlint, oxfmt, markers, configs, scripts)"
	} else {
		log.Status = "warn"
		log.Message = fmt.Sprintf("%d Skip-readiness issue(s) — não bloqueia o ZIP mas pode quebrar sync no painel Skip", len(failures))
		log.Details = failures
	}
	log.Duration = time.Since(start).String()
	return log, nil
}

// runOxCheck runs `pnpm exec <tool> <args...>` inside OutputDir.
// Returns ("", true) on exit 0; ("<tail of stderr>", false) on non-zero.
// If the tool isn't installed at all, returns ("not installed", false).
func runOxCheck(dir, tool string, args []string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	full := append([]string{"exec", tool}, args...)
	cmd := exec.CommandContext(ctx, "pnpm", full...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err == nil {
		return "", true
	}
	msg := strings.TrimSpace(out.String())
	if msg == "" {
		msg = err.Error()
	}
	return truncate(msg, 400), false
}

// checkSkipProtectedMarkers verifies the two @skip-protected markers in index.html
// and the one in src/main.tsx are still in place.
func checkSkipProtectedMarkers(dir string) (string, bool) {
	indexPath := filepath.Join(dir, "index.html")
	mainPath := filepath.Join(dir, "src", "main.tsx")

	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		return "@skip-protected: index.html unreadable: " + err.Error(), false
	}
	if !bytes.Contains(indexData, []byte("@skip-protected")) ||
		!bytes.Contains(indexData, []byte("https://goskip.dev/skip.js")) {
		return "@skip-protected: index.html missing skip.js script or marker", false
	}

	mainData, err := os.ReadFile(mainPath)
	if err != nil {
		return "@skip-protected: src/main.tsx unreadable: " + err.Error(), false
	}
	if !bytes.Contains(mainData, []byte("@skip-protected")) {
		return "@skip-protected: src/main.tsx missing marker", false
	}
	return "", true
}

// checkPackageScripts verifies the package.json carries the canonical Skip scripts.
func checkPackageScripts(dir string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return "unreadable: " + err.Error(), false
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return "invalid JSON: " + err.Error(), false
	}
	required := []string{"dev", "build", "lint", "format"}
	missing := []string{}
	for _, k := range required {
		if pkg.Scripts[k] == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return "missing scripts: " + strings.Join(missing, ", "), false
	}
	return "", true
}
