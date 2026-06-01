package migrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Libs that Lovable projects typically use but Skip baseline lacks.
// We add them at minimum-supported versions for React 19; pnpm will resolve compatible.
var lovableExtraDeps = map[string]string{
	"@tanstack/react-query": "^5.83.0",
	"@hello-pangea/dnd":     "^18.0.1",
	"html2pdf.js":           "^0.14.0",
	"marked":                "^18.0.3",
}

// phase02Dependencies merges Lovable's extra deps into Skip's package.json.
// It never downgrades existing Skip deps.
func phase02Dependencies(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "02-dependencies"}

	pkgPath := filepath.Join(opts.OutputDir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return log, fmt.Errorf("reading package.json: %w", err)
	}

	// Decode preserving order is tricky with map[string]any. For simplicity, decode/encode
	// with stable ordering via json.MarshalIndent (Go sorts map keys alphabetically).
	var pkg map[string]any
	if err := json.Unmarshal(data, &pkg); err != nil {
		return log, fmt.Errorf("parsing package.json: %w", err)
	}

	deps, _ := pkg["dependencies"].(map[string]any)
	if deps == nil {
		deps = map[string]any{}
		pkg["dependencies"] = deps
	}

	added := []string{}
	for name, version := range lovableExtraDeps {
		if _, exists := deps[name]; exists {
			continue
		}
		deps[name] = version
		added = append(added, name)
	}

	// Also: if source's package.json has additional deps not in Skip, copy them at source's version
	// (e.g. project-specific libs). But never overwrite existing Skip versions.
	srcPkgPath := filepath.Join(opts.SourceDir, "package.json")
	if srcData, err := os.ReadFile(srcPkgPath); err == nil {
		var srcPkg map[string]any
		if err := json.Unmarshal(srcData, &srcPkg); err == nil {
			if srcDeps, ok := srcPkg["dependencies"].(map[string]any); ok {
				for name, version := range srcDeps {
					if _, exists := deps[name]; exists {
						continue
					}
					// Skip target-stack-defining libs we don't want from source
					if isStackDefining(name) {
						continue
					}
					deps[name] = version
					added = append(added, name+"(from-source)")
				}
			}
		}
	}

	out, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return log, err
	}
	out = append(out, '\n')
	if err := os.WriteFile(pkgPath, out, 0o644); err != nil {
		return log, err
	}

	log.Status = "ok"
	log.Message = fmt.Sprintf("added %d dep(s): %v", len(added), added)
	log.Duration = time.Since(start).String()
	return log, nil
}

// isStackDefining returns true for libs that Skip pins on its own (we never accept
// the source's version of these).
func isStackDefining(name string) bool {
	switch name {
	case "react", "react-dom", "react-router-dom", "vite", "zod",
		"@hookform/resolvers", "react-day-picker", "sonner", "vaul",
		"react-resizable-panels", "next-themes":
		return true
	}
	return false
}
